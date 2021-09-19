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

func TestGeneralParser(t *testing.T) {
	p, _ := parser.NewGeneralParser()
	content := `The task {{  $.node }} (enabled: {{$.node.enabled}}) has completed.
Yours, {{$.user[create]}}`
	err := p.Parse(content)
	fmt.Println(p.(*parser.GeneralParser).GetPlaceholders())
	require.Nil(t, err)
}

func TestGeneralParser_Parse(t *testing.T) {
	var err error
	t.Cleanup(cleanup)

	nodeId := primitive.NewObjectID()
	_, err = mongo.GetMongoCol("nodes").Insert(bson.M{
		"_id":     nodeId,
		"name":    "Test Node",
		"enabled": true,
		"settings": bson.M{
			"max_runners": 8,
		},
	})
	require.Nil(t, err)
	spiderId := primitive.NewObjectID()
	_, err = mongo.GetMongoCol("spiders").Insert(bson.M{
		"_id": spiderId,
	})
	require.Nil(t, err)
	_, err = mongo.GetMongoCol("spider_stats").Insert(bson.M{
		"_id":          spiderId,
		"result_count": 5000,
	})
	require.Nil(t, err)
	userId := primitive.NewObjectID()
	_, err = mongo.GetMongoCol("users").Insert(bson.M{
		"_id":      userId,
		"no":       1001,
		"username": "Test Username",
	})
	require.Nil(t, err)
	userIdUpdate := primitive.NewObjectID()
	_, err = mongo.GetMongoCol("users").Insert(bson.M{
		"_id":      userIdUpdate,
		"no":       1002,
		"username": "Test2 Username",
	})
	require.Nil(t, err)

	taskId := primitive.NewObjectID()
	task := bson.M{
		"_id":       taskId,
		"node_id":   nodeId,
		"spider_id": spiderId,
	}
	_, err = mongo.GetMongoCol("task_stats").Insert(bson.M{
		"_id":          taskId,
		"result_count": 100,
	})
	require.Nil(t, err)
	_, err = mongo.GetMongoCol("artifacts").Insert(bson.M{
		"_id": taskId,
		"_sys": bson.M{
			"create_uid": userId,
			"update_uid": userIdUpdate,
		},
	})

	p, _ := parser.NewGeneralParser()
	template := `The task on node {{  $.node.name }} (enabled: {{$.node.enabled}}, max_runners: {{$.node.settings.max_runners}}) has completed.
Task Result Count: {{ $.:task_stat.result_count }}
Spider Result Count: {{ $.spider:stat.result_count }}
Yours, {{$.user.username}} ({{$.user.no}}) and {{$.user[update].username}} ({{$.user[update].no}})`
	err = p.Parse(template)
	require.Nil(t, err)

	content, err := p.Render(task)
	require.Nil(t, err)
	require.Equal(t, `The task on node Test Node (enabled: true, max_runners: 8) has completed.
Task Result Count: 100
Spider Result Count: 5000
Yours, Test Username (1001) and Test2 Username (1002)`, content)
}

func cleanup() {
	_ = mongo.GetMongoCol("nodes").Delete(nil)
	_ = mongo.GetMongoCol("spiders").Delete(nil)
	_ = mongo.GetMongoCol("spider_stats").Delete(nil)
	_ = mongo.GetMongoCol("tasks").Delete(nil)
	_ = mongo.GetMongoCol("task_stats").Delete(nil)
	_ = mongo.GetMongoCol("users").Delete(nil)
}
