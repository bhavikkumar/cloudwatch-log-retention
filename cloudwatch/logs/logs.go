package logs

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
)

const (
	defaultRetentionPeriod = 14
)

type requestParameters struct {
	CloudWatchLog CloudWatchLog `json:"requestParameters"`
}

type CloudWatchLog struct {
	LogGroupName    string `json:"logGroupName"`
	RetentionPeriod int64  `json:"retentionInDays"`
}

func NewFromEvent(event events.CloudWatchEvent) (cwLog CloudWatchLog) {
	cwLog = CloudWatchLog{}
	cwLog.parseCloudWatchEvent(event)
	return
}

func RetentionPeriod() int64 {
	period, _ := strconv.ParseInt(os.Getenv("RETENTION_PERIOD"), 10, 64)
	validPeriods := []int64{1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1827, 3653}
	m := make(map[int64]bool)
	for i := 0; i < len(validPeriods); i++ {
		m[validPeriods[i]] = true
	}

	if _, ok := m[period]; ok {
		return period
	}
	log.Warnf("Using default retention period. RETENTION_PERIOD %v is invalid. Allowed values are %v", period, validPeriods)
	return defaultRetentionPeriod
}

func (cwLog *CloudWatchLog) parseCloudWatchEvent(event events.CloudWatchEvent) {
	if len(event.Detail) <= 0 {
		log.WithFields(log.Fields{"id": event.Version, "detailType": event.DetailType, "source": event.Source}).Warn("CloudWatch Event missing detail section")
		return
	}

	var requestParameters requestParameters
	err := json.Unmarshal(event.Detail, &requestParameters)
	if err != nil {
		log.WithFields(log.Fields{"detail": event.Detail}).Warn("Could not parse CloudWatch Event details", err)
		return
	}
	cwLog.LogGroupName = requestParameters.CloudWatchLog.LogGroupName
	cwLog.RetentionPeriod = requestParameters.CloudWatchLog.RetentionPeriod
}

func (cwLog *CloudWatchLog) UpdateRetentionPolicy(client cloudwatchlogsiface.CloudWatchLogsAPI, retentionPeriod int64) (err error) {
	input := &cloudwatchlogs.PutRetentionPolicyInput{
		LogGroupName:    &cwLog.LogGroupName,
		RetentionInDays: &retentionPeriod,
	}
	_, err = client.PutRetentionPolicy(input)
	return
}
