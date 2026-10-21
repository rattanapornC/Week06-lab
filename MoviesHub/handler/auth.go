package handler

import (
    "encoding/json"
    "errors"
    "net/http"
    "strings"
    // "unicode/utf8"

    "moviehub/auth"
    "moviehub/mw"
    "moviehub/store"
)

type loginReq struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

type registerReq struct {
    Email       string `json:"email"`
    Password    string `json:"password"`
    DisplayName string `json:"displayName"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
    var req loginReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "BAD_JSON", "อ่านกล่อง JSON ไม่ออก")
        return
    }
    email := strings.ToLower(strings.TrimSpace(req.Email))
    m, err := h.Store.MemberByEmail(r.Context(), email)
    if errors.Is(err, store.ErrNotFound) {
        writeError(w, http.StatusUnauthorized, "BAD_CREDENTIALS", "อีเมลหรือรหัสผ่านไม่ถูกต้อง")
        return
    }
    if err != nil {
        internal(w, err)
        return
    }
    if !auth.CheckPassword(m.PasswordHash, req.Password) {
        writeError(w, http.StatusUnauthorized, "BAD_CREDENTIALS", "อีเมลหรือรหัสผ่านไม่ถูกต้อง")
        return
    }
    token, err := h.Tickets.Issue(m.ID, m.Role)
    if err != nil {
        internal(w, err)
        return
    }
    writeJSON(w, http.StatusOK, map[string]any{"token": token, "member": m})
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
    var req registerReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "BAD_JSON", "อ่านกล่อง JSON ไม่ออก")
        return
    }
    email := strings.ToLower(strings.TrimSpace(req.Email))
    name := strings.TrimSpace(req.DisplayName)

    if !strings.Contains(email, "@") {
        writeError(w, http.StatusBadRequest, "BAD_EMAIL", "อีเมลไม่ถูกต้อง")
        return
    }
    if len(req.Password) < 4 {
        writeError(w, http.StatusBadRequest, "BAD_PASSWORD", "รหัสผ่านสั้นเกินไป")
        return
    }
    if name == "" {
        writeError(w, http.StatusBadRequest, "BAD_NAME", "กรุณากรอกชื่อที่แสดง")
        return
    }

    hash, err := auth.HashPassword(req.Password)
    if err != nil {
        internal(w, err)
        return
    }

    m, err := h.Store.CreateMember(r.Context(), email, hash, name)
    if errors.Is(err, store.ErrDuplicate) {
        writeError(w, http.StatusConflict, "EMAIL_TAKEN", "อีเมลนี้ถูกใช้แล้ว")
        return
    }
    if err != nil {
        internal(w, err)
        return
    }

    token, err := h.Tickets.Issue(m.ID, m.Role)
    if err != nil {
        internal(w, err)
        return
    }
    writeJSON(w, http.StatusCreated, map[string]any{"token": token, "member": m})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
    id, ok := mw.MemberID(r.Context())
    if !ok {
        writeError(w, http.StatusUnauthorized, "NO_TOKEN", "กรุณาเข้าสู่ระบบก่อน")
        return
    }
    m, err := h.Store.MemberByID(r.Context(), id)
    if errors.Is(err, store.ErrNotFound) {
        writeError(w, http.StatusUnauthorized, "MEMBER_GONE", "ไม่พบบัญชีนี้แล้ว กรุณาเข้าสู่ระบบใหม่")
        return
    }
    if err != nil {
        internal(w, err)
        return
    }
    writeJSON(w, http.StatusOK, m)
}