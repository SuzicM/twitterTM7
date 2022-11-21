package data

import (
	"fmt"
	"log"
	"os"

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
	// Create 'student' keyspace
	err = session.Query(
		fmt.Sprintf(`CREATE KEYSPACE IF NOT EXISTS %s
					WITH replication = {
						'class' : 'SimpleStrategy',
						'replication_factor' : %d
					}`, "student", 1)).Exec()
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
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (user_id UUID, tweet_title text, tweet_body text, tweet_id TIMEUUID PRIMARY KEY ((user_id), tweet_id)) WITH CLUSTERING ORDER BY (tweet_id ASC)`,
			"tweets_by_user")).Exec()
	if err != nil {
		sr.logger.Println(err)
	}
}

// TO DO
func (sr *TweetRepo) GetTweetsByUser(id string) (TweetsByUser, error) {
	scanner := sr.session.Query(`SELECT user_id, tweet_title, tweet_body, toTimestamp(tweet_id) FROM tweets_by_user WHERE user_id = ?`,
		id).Iter().Scanner()

	var tweets TweetsByUser
	for scanner.Next() {
		var tweet TweetByUser
		err := scanner.Scan(&tweet.UserId, &tweet.TweetTitle, &tweet.TweetBody)
		if err != nil {
			sr.logger.Println(err)
			return tweets, err
		}
	}
	if err := scanner.Err(); err != nil {
		sr.logger.Println(err)
		return tweets, err
	}
	return tweets, nil
}

func (sr *TweetRepo) InsertTweetByUser(userTweet *TweetByUser) error {
	tweetId := gocql.TimeUUID()
	err := sr.session.Query(
		`INSERT INTO tweets_by_user (user_id, tweet_title, tweet_body, tweet_id) 
		VALUES (?, ?, ?)`,
		userTweet.UserId, userTweet.TweetTitle, userTweet.TweetBody, tweetId).Exec()
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
