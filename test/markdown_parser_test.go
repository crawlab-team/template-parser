package test

import (
	"fmt"
	"github.com/crawlab-team/crawlab-db/mongo"
	"github.com/crawlab-team/template-parser"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"testing"
)

func TestMarkdownParser(t *testing.T) {
	p, _ := parser.NewGeneralParser()
	content := `The task {{  $.node }} (enabled: {{$.node.enabled}}) has completed.
Yours, {{$.user}}`
	err := p.Parse(content)
	fmt.Println(p.(*parser.GeneralParser).GetPlaceholders())
	require.Nil(t, err)
}

func TestMarkdownParser_Parse(t *testing.T) {
	var err error
	t.Cleanup(cleanup)

	nodeId := primitive.NewObjectID()
	_, err = mongo.GetMongoCol("nodes").Insert(bson.M{
		"_id":     nodeId,
		"name":    "Test Node",
		"enabled": true,
	})
	require.Nil(t, err)
	userId := primitive.NewObjectID()
	_, err = mongo.GetMongoCol("users").Insert(bson.M{
		"_id":      userId,
		"no":       1001,
		"username": "Test Username",
	})
	require.Nil(t, err)

	p, _ := parser.NewGeneralParser()
	template := `The task on node {{  $.node.name }} (enabled: {{$.node.enabled}}) has completed.
Yours, {{$.user.username}} (UserNo: {{$.user.no}})`
	err = p.Parse(template)
	require.Nil(t, err)

	task := bson.M{
		"node_id": nodeId,
		"user_id": userId,
	}
	content, err := p.Render(task)
	require.Nil(t, err)
	require.Equal(t, `The task on node Test Node (enabled: true) has completed.
Yours, Test Username (UserNo: 1001)`, content)
}

func cleanup() {
	_ = mongo.GetMongoCol("nodes").Delete(nil)
	_ = mongo.GetMongoCol("users").Delete(nil)
}
