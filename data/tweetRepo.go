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

func (sr *TweetRepo) InsertTweetByLike(userTweet *TweetByUsername) error {

	err := sr.session.Query(
		`INSERT INTO like_by_tweet (tweet_like, username, tweet_title, tweet_body, created_on) 
		VALUES (?, ?, ?, ?, ?)`,
		userTweet.TweetLike, userTweet.Username, userTweet.TweetTitle, userTweet.TweetBody, userTweet.CreatedOn).Exec()
	if err != nil {
		sr.logger.Println(err)
		return err
	}
	return nil
}
func (sr *TweetRepo) UpdateLikeByTweet(username string, tweetBody string, tweetLike string) error {

	err := sr.session.Query(
		`UPDATE like_by_tweet SET tweetLike=tweetLike+? where username = ? and tweetBody = ? `,
		[]string{tweetLike}, username, tweetBody).Exec()
	if err != nil {
		sr.logger.Println(err)
		return err
	}
	return nil
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
