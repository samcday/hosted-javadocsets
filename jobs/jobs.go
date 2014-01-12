package jobs

import (
	"github.com/garyburd/redigo/redis"
	"github.com/samcday/hosted-javadocsets/redisconn"
)

// QueueDocsetJob will queue a job to build a docset for an artifact, if there
// is not yet one built.
func QueueDocsetJob(groupId, artifactId string, version string) error {
	var redisConn redis.Conn = redisconn.Get()

	id := groupId + ":" + artifactId

	exists, err := redis.Bool(redisConn.Do("SISMEMBER", "docsets", id))
	if err != nil {
		return err
	}
	if exists == true && version != "" {
		verExists, err := redis.Bool(redisConn.Do("SISMEMBER", "docset:"+id, version))
		if err != nil || verExists {
			return err
		}
	} else if exists == true {
		return nil
	}

	if err := QueueJob(map[string]string{
		"Job":        "build-docset",
		"ArtifactId": artifactId,
		"GroupId":    groupId,
		"Version":    version,
	}); err != nil {
		return err
	}

	return nil
}
