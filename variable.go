package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/crawlab-team/crawlab-db/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strings"
)

type Variable struct {
	root   interface{}
	tokens []string
	doc    bson.M
}

func (v *Variable) GetValue() (value interface{}, err error) {
	return v.getNodeByIndex(len(v.tokens) - 1)
}

func (v *Variable) getNodeByIndex(index int) (result interface{}, err error) {
	node := v.doc
	for i := 0; i < index && i < len(v.tokens); i++ {
		nextIndex := i + 1
		if nextIndex < len(v.tokens)-1 {
			// root or intermediate node
		} else {
			// value
			return v.getNextValue(node, i), nil
		}

		// next node
		node, err = v.getNextNode(node, i)
		if err != nil {
			return nil, err
		}
		if node == nil {
			return nil, nil
		}
	}
	return node, nil
}

func (v *Variable) getNextNode(currentNode bson.M, currentIndex int) (nextNode bson.M, err error) {
	// next index and token
	nextIndex := currentIndex + 1
	nextToken := v.tokens[nextIndex]

	// next id
	nextIdKey := fmt.Sprintf("%s_id", nextToken)
	nextIdRes, ok := currentNode[nextIdKey]
	if !ok {
		return nil, errors.New(fmt.Sprintf("%s is not available in %s", nextIdKey, strings.Join(v.tokens[:nextIndex], ".")))
	}
	nextId, ok := nextIdRes.(primitive.ObjectID)
	if !ok {
		nextIdStr, ok := nextIdRes.(string)
		if !ok {
			return nil, errors.New(fmt.Sprintf("%s is not ObjectId in %s", nextIdKey, strings.Join(v.tokens[:nextIndex], ".")))
		}
		nextId, err = primitive.ObjectIDFromHex(nextIdStr)
		if err != nil {
			return nil, err
		}
	}
	if nextId.IsZero() {
		return nil, nil
	}

	// mongo collection name
	colName := fmt.Sprintf("%ss", nextToken)

	// get next node from mongo collection
	if err := mongo.GetMongoCol(colName).FindId(nextId).One(&nextNode); err != nil {
		return nil, err
	}

	return nextNode, nil
}

func (v *Variable) getNextValue(currentNode bson.M, currentIndex int) (nextValue interface{}) {
	// next index and token
	nextIndex := currentIndex + 1
	nextToken := v.tokens[nextIndex]

	// next value
	nextValue, _ = currentNode[nextToken]

	return nextValue
}

func NewVariable(root interface{}, placeholder string) (v *Variable, err error) {
	// validate
	if placeholder == "" {
		return nil, errors.New("empty placeholder")
	}
	if !strings.HasPrefix(placeholder, "$") {
		return nil, errors.New("not start with $")
	}

	// tokens
	tokens := strings.Split(placeholder, ".")

	// document
	data, err := json.Marshal(root)
	if err != nil {
		return nil, err
	}
	var doc bson.M
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}

	v = &Variable{
		root:   root,
		tokens: tokens,
		doc:    doc,
	}

	return v, nil
}
