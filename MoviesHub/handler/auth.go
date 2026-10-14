package handler

import (
    "encoding/json"
    "errors"
    "net/http"
    "strings"
    // "unicode/utf8"

    "moviehub/auth"
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