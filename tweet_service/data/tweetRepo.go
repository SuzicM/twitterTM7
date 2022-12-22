package data

import (
	"fmt"
	"log"
	"os"
	"strings"

	// NoSQL: module containing Cassandra api client
	"github.com/gocql/gocql"
)

// NoSQL: StudentRepo struct encapsulating Cassandra api client
type TweetRepo struct {
	session *gocql.Session
	logger  *log.Logger
}

// NoSQL: Constructor which reads db configuration from environment and creates a keyspace
func New(logger *log.Logger) (*TweetRepo, error) {
	db := os.Getenv("CASS_DB")

	// Connect to default keyspace
	cluster := gocql.NewCluster(db)
	cluster.Keyspace = "system"
	session, err := cluster.CreateSession()
	if err != nil {
		logger.Println(err)
		return nil, err
	}
	// Create 'tweet' keyspace
	err = session.Query(
		fmt.Sprintf(`CREATE KEYSPACE IF NOT EXISTS %s
					WITH replication = {
						'class' : 'SimpleStrategy',
						'replication_factor' : %d
					}`, "tweet", 1)).Exec()
	if err != nil {
		logger.Println(err)
	}
	session.Close()

	// Connect to tweet keyspace
	cluster.Keyspace = "tweet"
	cluster.Consistency = gocql.One
	session, err = cluster.CreateSession()
	if err != nil {
		logger.Println(err)
		return nil, err
	}

	// Return repository with logger and DB session
	return &TweetRepo{
		session: session,
		logger:  logger,
	}, nil
}

// Disconnect from database
func (sr *TweetRepo) CloseSession() {
	sr.session.Close()
}

// Create tables
func (sr *TweetRepo) CreateTables() {
	err := sr.session.Query(
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s 
		(user_id UUID, tweet_title text, tweet_body text, created_on TIMEUUID,
		PRIMARY KEY ((user_id), created_on)) 
		WITH CLUSTERING ORDER BY (created_on ASC)`,
			"tweets_by_user")).Exec()
	if err != nil {
		sr.logger.Println(err)
	}

	err = sr.session.Query(
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s 
		(username text, tweet_title text, tweet_body text, created_on TIMEUUID,
		PRIMARY KEY ((username), created_on)) 
		WITH CLUSTERING ORDER BY (created_on ASC)`,
			"tweets_by_username")).Exec()
	if err != nil {
		sr.logger.Println(err)
	}

	err = sr.session.Query(
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s 
		(username text, tweetid TIMEUUID, liked boolean,
		PRIMARY KEY ((username), tweetid)) 
		WITH CLUSTERING ORDER BY (tweetid ASC)`,
			"user_likes")).Exec()
	if err != nil {
		sr.logger.Println(err)
	}
}

// TO DO
func (sr *TweetRepo) GetTweetsByUser(id string) (TweetsByUser, error) {
	scanner := sr.session.Query(`SELECT user_id, tweet_title, tweet_body, created_on FROM tweets_by_user WHERE user_id = ?`,
		id).Iter().Scanner()

	var tweets TweetsByUser
	for scanner.Next() {
		var tweet TweetByUser
		err := scanner.Scan(&tweet.UserId, &tweet.TweetTitle, &tweet.TweetBody, &tweet.CreatedOn)
		if err != nil {
			sr.logger.Println(err)
			return tweets, err
		}
		userTweetChanged := tweet.TweetBody
		userTweetChanged = strings.ReplaceAll(userTweetChanged, "i16", "<")
		userTweetChanged = strings.ReplaceAll(userTweetChanged, "i12", ">")
		tweet.TweetBody = userTweetChanged
		tweets = append(tweets, &tweet)
	}
	if err := scanner.Err(); err != nil {
		sr.logger.Println(err)
		return tweets, err
	}
	return tweets, nil
}

func (sr *TweetRepo) GetTweetsByUsername(id string) (TweetsByUsername, error) {
	scanner := sr.session.Query(`SELECT username, tweet_title, tweet_body, created_on FROM tweets_by_username WHERE username = ?`,
		id).Iter().Scanner()

	var tweets TweetsByUsername
	for scanner.Next() {
		var tweet TweetByUsername
		err := scanner.Scan(&tweet.Username, &tweet.TweetTitle, &tweet.TweetBody, &tweet.CreatedOn)
		if err != nil {
			sr.logger.Println(err)
			return tweets, err
		}
		userTweetChanged := tweet.TweetBody
		userTweetChanged = strings.ReplaceAll(userTweetChanged, "i16", "<")
		userTweetChanged = strings.ReplaceAll(userTweetChanged, "i12", ">")
		tweet.TweetBody = userTweetChanged
		tweets = append(tweets, &tweet)
	}
	if err := scanner.Err(); err != nil {
		sr.logger.Println(err)
		return tweets, err
	}
	return tweets, nil
}

func (sr *TweetRepo) GetUserLikes(id gocql.UUID) (int, error) {
	scanner := sr.session.Query(`SELECT username, tweetid, liked FROM user_likes WHERE tweetid = ? ALLOW FILTERING`,
		id).Iter().Scanner()

	likes := 0
	for scanner.Next() {
		var like Like
		err := scanner.Scan(&like.Username, &like.TweetId, &like.Liked)
		if err != nil {
			sr.logger.Println("error while getting likes, scanner")
			sr.logger.Println(err)
			return 0, err
		}
		if like.Liked {
			likes += 1
		}
	}
	if err := scanner.Err(); err != nil {
		sr.logger.Println("error while getting likes, after scanner")
		sr.logger.Println(err)
		return 0, err
	}
	return likes, nil
}

func (sr *TweetRepo) GetUsersThatLiked(id gocql.UUID) (UsersLiked, error) {
	scanner := sr.session.Query(`SELECT username, tweetid, liked FROM user_likes WHERE tweetid = ? ALLOW FILTERING`,
		id).Iter().Scanner()

	var users []string
	for scanner.Next() {
		var like Like
		err := scanner.Scan(&like.Username, &like.TweetId, &like.Liked)
		if err != nil {
			sr.logger.Println("error while getting list of users, scanner")
			sr.logger.Println(err)
			return nil, err
		}
		if like.Liked {
			users = append(users, like.Username)
		}
	}
	if err := scanner.Err(); err != nil {
		sr.logger.Println("error while getting list of users, after scanner")
		sr.logger.Println(err)
		return nil, err
	}
	return users, nil
}

func (sr *TweetRepo) InsertTweetByUser(userTweet *TweetByUser) error {
	userid, _ := gocql.RandomUUID()
	created := gocql.TimeUUID()
	err := sr.session.Query(
		`INSERT INTO tweets_by_user (user_id, tweet_title, tweet_body, created_on) 
		VALUES (?, ?,  ?, ?)`,
		userid, userTweet.TweetTitle, userTweet.TweetBody, created).Exec()
	if err != nil {
		sr.logger.Println(err)
		return err
	}
	return nil
}

func (sr *TweetRepo) InsertTweetByUsername(userTweet *TweetByUsername) error {
	created := gocql.TimeUUID()
	err := sr.session.Query(
		`INSERT INTO tweets_by_username (username, tweet_title, tweet_body, created_on) 
		VALUES ( ?, ?, ?, ?)`,
		userTweet.Username, userTweet.TweetTitle, userTweet.TweetBody, created).Exec()
	if err != nil {
		sr.logger.Println(err)
		return err
	}
	return nil
}

func (sr *TweetRepo) LikeDislikeTweet(username string, id gocql.UUID) (bool, error) {

	newLike := Like{}
	newLike.Username = username
	newLike.TweetId = id

	scanner := sr.session.Query(`SELECT username, tweetid, liked FROM user_likes WHERE username = ? AND tweetid = ? ALLOW FILTERING`,
		newLike.Username, newLike.TweetId).Iter().Scanner()

	var existLike Like // dobije lajk za dani username i tweetid
	for scanner.Next() {
		err := scanner.Scan(&existLike.Username, &existLike.TweetId, &existLike.Liked)
		if err != nil {
			sr.logger.Println("error while checking like, scanner")
			sr.logger.Println(err)
		}
	}
	if err := scanner.Err(); err != nil {
		sr.logger.Println("error while checking like, after scanner")
		sr.logger.Println(newLike.Username)
		sr.logger.Println(err)
	}

	newLike.Liked = !existLike.Liked // ako je bio lajk sad je nije lajkovan i obrnuto

	err := sr.session.Query(
		`UPDATE user_likes SET liked = ? WHERE tweetid = ? AND username = ?`,
		newLike.Liked, newLike.TweetId, newLike.Username,).Exec()
	if err != nil {
		sr.logger.Println("Error updating like")
		sr.logger.Println(newLike.TweetId)
		sr.logger.Println(err)
	} else {
		return true, nil //uspjesna promjena iz lajkovanog u ne lajkovani ili obrnuto
	}

	//dodavanje novog ako tweet nije uopste imao interakciju korisnika
	err = sr.session.Query(
		`INSERT INTO user_likes (username, tweetid, liked) 
		VALUES ( ?, ?, ?)`,
		newLike.Username, newLike.TweetId, newLike.Liked).Exec()
	if err != nil {
		sr.logger.Println("error adding like")
		sr.logger.Println(newLike.Username)
		sr.logger.Println(err)
		return false, err
	}

	return true, nil
}

// NoSQL: Performance issue, we never want to fetch all the data
func (sr *TweetRepo) GetDistinctIds(idColumnName string, tableName string) ([]string, error) {
	scanner := sr.session.Query(
		fmt.Sprintf(`SELECT DISTINCT %s FROM %s`, idColumnName, tableName)).
		Iter().Scanner()
	var ids []string
	for scanner.Next() {
		var id string
		err := scanner.Scan(&id)
		if err != nil {
			sr.logger.Println(err)
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := scanner.Err(); err != nil {
		sr.logger.Println(err)
		return nil, err
	}
	return ids, nil
}
