package api

import (
	"encoding/json"
	"net/http"

	errs "github.com/ONSdigital/dp-census-alpha-search-api/apierrors"
	"github.com/ONSdigital/dp-census-alpha-search-api/models"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
)

func (api *SearchAPI) getTaxonomy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	setAccessControl(w, http.MethodGet)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	log.Event(ctx, "getTaxonomy endpoint: incoming request", log.INFO)

	b, err := json.Marshal(api.taxonomy)
	if err != nil {
		log.Event(ctx, "getTaxonomy endpoint: failed to marshal taxonomy resource into bytes", log.ERROR, log.Error(err))
		setErrorCode(w, errs.ErrInternalServer)
	}

	_, err = w.Write(b)
	if err != nil {
		log.Event(ctx, "getTaxonomy endpoint: error writing response", log.ERROR, log.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	log.Event(ctx, "getTaxonomy endpoint: successfully retreived taxonomy", log.INFO)
}

// Topic ...
type Topic struct {
	ParentTopic string   `json:"parent_topic,omitempty"`
	Title       string   `json:"title"`
	Topic       string   `json:"topic"`
	ChildTopics []string `json:"child_topics,omitempty"`
}

func (api *SearchAPI) getTopic(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	setAccessControl(w, http.MethodGet)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	vars := mux.Vars(r)
	topic := vars["topic"]
	logData := log.Data{"topic": topic}

	log.Event(ctx, "getTopic endpoint: incoming request", log.INFO, logData)

	var result *Topic
	var hasValidTopic bool
	for _, taxonomy := range api.taxonomy.Topics {
		if topic == taxonomy.FormattedTitle {
			hasValidTopic = true

			var childTopics []string
			for _, childTopic := range taxonomy.ChildTopics {
				childTopics = append(childTopics, childTopic.FormattedTitle)
			}

			result = &Topic{
				Title:       taxonomy.Title,
				Topic:       topic,
				ChildTopics: childTopics,
			}

			break
		}

		result, hasValidTopic = checkChildTopics(taxonomy, topic)
		if hasValidTopic {
			break
		}
	}

	if !hasValidTopic {
		err := errs.ErrTopicNotFound
		log.Event(ctx, "getTopic endpoint: failed to find topic", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}

	b, err := json.Marshal(result)
	if err != nil {
		log.Event(ctx, "getTopic endpoint: failed to marshal topic resource into bytes", log.ERROR, log.Error(err), logData)
		setErrorCode(w, errs.ErrInternalServer)
	}

	_, err = w.Write(b)
	if err != nil {
		log.Event(ctx, "getTopic endpoint: error writing response", log.ERROR, log.Error(err), logData)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	log.Event(ctx, "getTopic endpoint: successfully retrieved topic", log.INFO, logData)
}

func checkChildTopics(taxonomy models.Topic, topic string) (result *Topic, hasValidTopic bool) {

	for _, childTopic := range taxonomy.ChildTopics {
		if topic == childTopic.FormattedTitle {
			hasValidTopic = true

			var childTopics []string
			for _, childTopic := range childTopic.ChildTopics {
				childTopics = append(childTopics, childTopic.FormattedTitle)
			}

			result = &Topic{
				ParentTopic: taxonomy.FormattedTitle,
				Title:       childTopic.Title,
				Topic:       topic,
				ChildTopics: childTopics,
			}

			break
		}

		result, hasValidTopic = checkChildTopics(childTopic, topic)
		if hasValidTopic {
			break
		}
	}

	return
}
