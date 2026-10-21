package store

import (
    "context"
    "errors"

    "github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5"
)

var ErrEmailTaken = errors.New("email taken")

type Member struct {
    ID           int64  `json:"id"`
    Email        string `json:"email"`
    DisplayName  string `json:"displayName"`
    Role         string `json:"role"`
    PasswordHash string `json:"-"` // ไม่มีวันออกไปกับ JSON
}

func (s *Store) MemberByEmail(ctx context.Context, email string) (Member, error) {
    var m Member
    err := s.Pool.QueryRow(ctx,
        `SELECT id, email, display_name, role, password_hash
         FROM members WHERE lower(email) = $1`, email).
        Scan(&m.ID, &m.Email, &m.DisplayName, &m.Role, &m.PasswordHash)
    if errors.Is(err, pgx.ErrNoRows) {
        return Member{}, ErrNotFound
    }
    return m, err
}

var ErrDuplicate = errors.New("duplicate")

func (s *Store) CreateMember(ctx context.Context, email, passwordHash, displayName string) (Member, error) {
    var m Member
    err := s.Pool.QueryRow(ctx, `
        INSERT INTO members (email, password_hash, display_name)
        VALUES ($1, $2, $3)
        RETURNING id, email, password_hash, display_name, role`,
        email, passwordHash, displayName,
    ).Scan(&m.ID, &m.Email, &m.PasswordHash, &m.DisplayName, &m.Role)
    if err != nil {
        var pgErr *pgconn.PgError
        if errors.As(err, &pgErr) && pgErr.Code == "23505" {
            return Member{}, ErrDuplicate
        }
        return Member{}, err
    }
    return m, nil
}