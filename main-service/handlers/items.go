package handlers

import (
	"encoding/json"
	"main-service/cache"
	"main-service/events"
	"main-service/store"
	"net/http"
)

type ItemHandler interface {
	CreateItem(w http.ResponseWriter, r *http.Request)
}

type itemHandler struct {
	store    store.Store
	cache    cache.Cache
	producer events.Producer
}

func NewItemHandler(
	s store.Store,
	c cache.Cache,
	p events.Producer,
) ItemHandler {
	return &itemHandler{
		s,
		c,
		p,
	}
}

func (h *itemHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	id, err := h.store.CreateItem(r.Context(), req.Name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_ = h.cache.Set(r.Context(), "item:"+req.Name, "created")
	_ = h.producer.Publish(r.Context(), "items.created", []byte(req.Name))
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"id": id, "name": req.Name})
}
