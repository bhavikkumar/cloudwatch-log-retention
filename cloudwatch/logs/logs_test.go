package logs_test

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
	"github.com/bhavikkumar/cloudwatch-log-retention/cloudwatch/logs"
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"testing"
)

func TestDefaultRetentionPeriod(t *testing.T) {
	assert.Equal(t, logs.RetentionPeriod(), int64(14))
}

func TestInvalidRetetionPeriod(t *testing.T) {
	os.Setenv("RETENTION_PERIOD", "2")
	assert.Equal(t, logs.RetentionPeriod(), int64(14))
}

func TestValidRetentionPeriod(t *testing.T) {
	validPeriods := []int64{1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1827, 3653}
	for _, value := range validPeriods {
		os.Setenv("RETENTION_PERIOD", strconv.FormatInt(value, 10))
		actual := logs.RetentionPeriod()
		assert.Equal(t, value, actual)
	}
}

func TestNewCloudWatchLogFromEventWithNoRequestDetail(t *testing.T) {
	eventJson := `{"version": "0","id": "560646ad-c1f3-a8fb-ea6d-730c1b2bfd63","detail-type": "AWS API Call via CloudTrail","source": "aws.logs","account": "336840772780","time": "2019-04-10T19:18:47Z","region": "us-east-1","resources": []}`
	var event events.CloudWatchEvent
	json.Unmarshal([]byte(eventJson), &event)
	actual := logs.NewFromEvent(event)
	assert.Empty(t, actual.LogGroupName)
	assert.Empty(t, actual.RetentionPeriod)
}

func TestNewCloudWatchLogFromEventNoRequestParameters(t *testing.T) {
	eventJson := `{"version": "0","id": "560646ad-c1f3-a8fb-ea6d-730c1b2bfd63","detail-type": "AWS API Call via CloudTrail","source": "aws.logs","account": "336840772780","time": "2019-04-10T19:18:47Z","region": "us-east-1","resources": [], "detail": { "requestParameters":{}}}`
	var event events.CloudWatchEvent
	json.Unmarshal([]byte(eventJson), &event)
	actual := logs.NewFromEvent(event)
	assert.Empty(t, actual.LogGroupName)
	assert.Empty(t, actual.RetentionPeriod)
}

func TestNewCloudWatchLogFromMalformedEvent(t *testing.T) {
	eventJson := `{"version": "0","id": "560646ad-c1f3-a8fb-ea6d-730c1b2bfd63","detail-type": "AWS API Call via CloudTrail","source": "aws.logs","account": "336840772780","time": "2019-04-10T19:18:47Z","region": "us-east-1","resources": [], "detail": { "requestParameters": { "logGroupName": 1}}}`
	var event events.CloudWatchEvent
	json.Unmarshal([]byte(eventJson), &event)
	actual := logs.NewFromEvent(event)
	assert.Empty(t, actual.LogGroupName)
	assert.Empty(t, actual.RetentionPeriod)
}

func TestNewCloudWatchLogFromEvent(t *testing.T) {
	eventJson := `{"version": "0","id": "560646ad-c1f3-a8fb-ea6d-730c1b2bfd63","detail-type": "AWS API Call via CloudTrail","source": "aws.logs","account": "336840772780","time": "2019-04-10T19:18:47Z","region": "us-east-1","resources": [], "detail": { "requestParameters": { "logGroupName": "test", "retentionInDays": 7}}}`
	var event events.CloudWatchEvent
	json.Unmarshal([]byte(eventJson), &event)
	actual := logs.NewFromEvent(event)
	expected := logs.CloudWatchLog{LogGroupName: "test", RetentionPeriod: 7}
	assert.Equal(t, expected, actual)
}

func TestUpdateRetentionPolicy(t *testing.T) {
	cloudWatchLog := logs.CloudWatchLog{LogGroupName: "test", RetentionPeriod: 7}
	err := cloudWatchLog.UpdateRetentionPolicy(mockedPutRetentionPolicy{}, 7)
	assert.NoError(t, err)
}

type mockedPutRetentionPolicy struct {
	cloudwatchlogsiface.CloudWatchLogsAPI
	Resp cloudwatchlogs.PutRetentionPolicyOutput
}

func (m mockedPutRetentionPolicy) PutRetentionPolicy(in *cloudwatchlogs.PutRetentionPolicyInput) (*cloudwatchlogs.PutRetentionPolicyOutput, error) {
	return &m.Resp, nil
}
