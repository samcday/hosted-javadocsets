package main

import (
	"bytes"
	"io"
	"strings"
	"time"

	log "github.com/cihub/seelog"
	"github.com/crowdmob/goamz/s3"
	"github.com/joho/godotenv"
	"github.com/kr/beanstalk"
	"github.com/samcday/hosted-javadocsets/awsutil"
	"github.com/samcday/hosted-javadocsets/docset"
	"github.com/samcday/hosted-javadocsets/jobs"
)

func main() {
	defer log.Flush()

	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	// bucket := awsutil.Bucket()
	// log.Warn(bucket.Name)
	// data := []byte("Hello, Goamz!!")
	// if err := bucket.Put("sample.txt", data, "text/plain", s3.BucketOwnerFull, s3.Options{}); err != nil {
	// 	panic(err)
	// }
	// if 1 == 1 {
	// 	return
	// }

	for {
		job, err := jobs.TakeJob(60 * time.Second)
		if err != nil {
			if connErr, ok := err.(beanstalk.ConnError); ok && connErr.Err == beanstalk.ErrTimeout {
				time.Sleep(1 * time.Second)
				continue
			}
			// Crash driven design. If we got an error taking a job and it
			// wasn't because we timed out waiting for one, we crash out and
			// hope a fresh instance has better luck. Of course if beanstalk
			// is down we're gonna keep crashing, but our supervisor should be
			// doing exponential backoff on process restarts anyway.
			panic(err)
		}

		payload := job.Payload()
		jobName := payload["Job"]
		log.Infof("Processing job %s", jobName)
		log.Debug("Job data ", payload)
		switch jobName {
		case "build-docset":
			// TODO: some kind of lock here in case we queue multiple jobs to
			// build the same docset.
			err = buildDocset(payload["GroupId"], payload["ArtifactId"], payload["Version"])
		}

		if err != nil {
			log.Warnf("Job %s failed", jobName, err)
			job.Release(1 * time.Minute)
		} else {
			log.Infof("Job %s was successful.", jobName)
			job.Complete()
		}
	}
}

func buildDocset(groupId, artifactId string, version string) error {
	// Okay. This makes me kinda love go standard library. Just like Node!
	reader, writer := io.Pipe()

	go func() {
		docset.Create(groupId, artifactId, version, writer)
	}()
	key := strings.Replace(groupId, ".", "/", -1) + "/" + artifactId + "-" + version + ".tgz"
	return chunkedS3Upload(key, "application/x-gzip", s3.PublicRead, reader)
}

func chunkedS3Upload(key, contentType string, acl s3.ACL, reader io.Reader) error {
	log.Debug("Starting multipart upload to S3.")
	bucket := awsutil.Bucket()
	multi, err := bucket.Multi(key, contentType, acl)
	if err != nil {
		panic(err)
		return err
	}

	data := make([]byte, 1024*1024*5)
	parts := make([]s3.Part, 0)
	eof := false

	for eof == false {
		view := data
		totalRead := 0
		for totalRead < len(data) {
			read, err := reader.Read(view)
			view = view[read:]
			totalRead += read
			if err != nil && err == io.EOF {
				eof = true
				break
			} else if err != nil {
				multi.Abort()
				return err
			}
		}
		part, err := multi.PutPart(len(parts)+1, bytes.NewReader(data[0:totalRead]))
		if err != nil {
			multi.Abort()
			return err
		}
		parts = append(parts, part)
	}

	log.Debugf("S3 multipart uploaded all chunks successfully. Completing upload...")
	return multi.Complete(parts)
}
