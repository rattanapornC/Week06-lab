package handler

import (
    "encoding/json"
    "errors"
    "net/http"
    "strings"
    "unicode/utf8"

    "moviehub/mw"
    "moviehub/store"
)

func (h *Handler) PostReview(w http.ResponseWriter, r *http.Request) {
    movieID, ok := pathID(r, "id")
    if !ok {
        writeError(w, http.StatusBadRequest, "BAD_ID", "เลขหนังต้องเป็นตัวเลข")
        return
    }
    memberID, _ := mw.MemberID(r.Context())

    var req struct {
        Text string `json:"text"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "BAD_JSON", "อ่านกล่อง JSON ไม่ออก")
        return
    }
    text := strings.TrimSpace(req.Text)
    if utf8.RuneCountInString(text) < 10 {
        writeError(w, http.StatusBadRequest, "BAD_TEXT", "รีวิวต้องยาวอย่างน้อย 10 ตัวอักษร")
        return
    }

    rev, err := h.Store.CreateReview(r.Context(), memberID, movieID, text)
    if errors.Is(err, store.ErrBadMovie) {
        writeError(w, http.StatusNotFound, "MOVIE_NOT_FOUND", "ไม่พบหนังเรื่องนี้")
        return
    }
    if err != nil {
        internal(w, err)
        return
    }
    writeJSON(w, http.StatusCreated, rev)
}

func (h *Handler) PutVote(w http.ResponseWriter, r *http.Request) {
    movieID, ok := pathID(r, "id")
    if !ok {
        writeError(w, http.StatusBadRequest, "BAD_ID", "เลขหนังต้องเป็นตัวเลข")
        return
    }
    memberID, _ := mw.MemberID(r.Context())

    var req struct {
        Score int `json:"score"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "BAD_JSON", "อ่านกล่อง JSON ไม่ออก")
        return
    }
    if req.Score < 1 || req.Score > 10 {
        writeError(w, http.StatusBadRequest, "BAD_SCORE", "คะแนนต้องอยู่ระหว่าง 1 ถึง 10")
        return
    }

    err := h.Store.SetVote(r.Context(), memberID, movieID, req.Score)
    if errors.Is(err, store.ErrBadMovie) {
        writeError(w, http.StatusNotFound, "MOVIE_NOT_FOUND", "ไม่พบหนังเรื่องนี้")
        return
    }
    if err != nil {
        internal(w, err)
        return
    }
    writeJSON(w, http.StatusOK, map[string]any{"score": req.Score})
}

func (h *Handler) ListWishlist(w http.ResponseWriter, r *http.Request) {
    memberID, _ := mw.MemberID(r.Context())
    items, err := h.Store.ListWishlist(r.Context(), memberID)
    if err != nil {
        internal(w, err)
        return
    }
    writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) AddWishlist(w http.ResponseWriter, r *http.Request) {
    movieID, ok := pathID(r, "id")
    if !ok {
        writeError(w, http.StatusBadRequest, "BAD_ID", "เลขหนังต้องเป็นตัวเลข")
        return
    }
    memberID, _ := mw.MemberID(r.Context())
    err := h.Store.AddWishlist(r.Context(), memberID, movieID)
    if errors.Is(err, store.ErrBadMovie) {
        writeError(w, http.StatusNotFound, "MOVIE_NOT_FOUND", "ไม่พบหนังเรื่องนี้")
        return
    }
    if err != nil {
        internal(w, err)
        return
    }
    w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RemoveWishlist(w http.ResponseWriter, r *http.Request) {
    movieID, ok := pathID(r, "id")
    if !ok {
        writeError(w, http.StatusBadRequest, "BAD_ID", "เลขหนังต้องเป็นตัวเลข")
        return
    }
    memberID, _ := mw.MemberID(r.Context())
    if err := h.Store.RemoveWishlist(r.Context(), memberID, movieID); err != nil {
        internal(w, err)
        return
    }
    w.WriteHeader(http.StatusNoContent)
}