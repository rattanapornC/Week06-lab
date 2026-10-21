package store

import (
    "context"
    "errors"

    "github.com/jackc/pgx/v5/pgconn"
)

var ErrBadMovie = errors.New("movie not found")

func isFKViolation(err error) bool {
    var pgErr *pgconn.PgError
    return errors.As(err, &pgErr) && pgErr.Code == "23503"
}

// ---------- รีวิว ----------
func (s *Store) CreateReview(ctx context.Context, memberID, movieID int64, text string) (Review, error) {
    r := Review{MemberID: memberID, MovieID: movieID, Text: text}
    err := s.Pool.QueryRow(ctx,
        `INSERT INTO reviews (member_id, movie_id, text)
         VALUES ($1, $2, $3)
         RETURNING id, created_at, updated_at`,
        memberID, movieID, text).Scan(&r.ID, &r.CreatedAt, &r.UpdatedAt)
    if err != nil {
        if isFKViolation(err) {
            return Review{}, ErrBadMovie
        }
        return Review{}, err
    }
    err = s.Pool.QueryRow(ctx,
        `SELECT display_name FROM members WHERE id = $1`, memberID).Scan(&r.Member.DisplayName)
    if err != nil {
        return Review{}, err
    }
    r.Member.ID = memberID
    s.Pool.QueryRow(ctx,
        `SELECT score FROM votes WHERE member_id = $1 AND movie_id = $2`,
        memberID, movieID).Scan(&r.Score) // ไม่มีโหวตก็ปล่อยเป็น nil
    return r, nil
}

// ---------- คะแนน ----------
func (s *Store) SetVote(ctx context.Context, memberID, movieID int64, score int) error {
    tag, err := s.Pool.Exec(ctx,
        `UPDATE votes SET score = $3 WHERE member_id = $1 AND movie_id = $2`,
        memberID, movieID, score)
    if err != nil {
        return err
    }
    if tag.RowsAffected() > 0 {
        return nil
    }
    _, err = s.Pool.Exec(ctx,
        `INSERT INTO votes (member_id, movie_id, score) VALUES ($1, $2, $3)`,
        memberID, movieID, score)
    if isFKViolation(err) {
        return ErrBadMovie
    }
    return err
}

// ---------- อยากดู ----------
func (s *Store) AddWishlist(ctx context.Context, memberID, movieID int64) error {
    _, err := s.Pool.Exec(ctx,
        `INSERT INTO wishlist_items (member_id, movie_id)
         SELECT $1, $2
         WHERE NOT EXISTS (
             SELECT 1 FROM wishlist_items WHERE member_id = $1 AND movie_id = $2)`,
        memberID, movieID)
    if isFKViolation(err) {
        return ErrBadMovie
    }
    return err
}

func (s *Store) RemoveWishlist(ctx context.Context, memberID, movieID int64) error {
    _, err := s.Pool.Exec(ctx,
        `DELETE FROM wishlist_items WHERE member_id = $1 AND movie_id = $2`,
        memberID, movieID)
    return err
}

func (s *Store) ListWishlist(ctx context.Context, memberID int64) ([]MovieDetail, error) {
    rows, err := s.Pool.Query(ctx,
        `SELECT movie_id FROM wishlist_items WHERE member_id = $1 ORDER BY movie_id DESC`, memberID)
    if err != nil {
        return nil, err
    }
    var ids []int64
    for rows.Next() {
        var id int64
        if err := rows.Scan(&id); err != nil {
            rows.Close()
            return nil, err
        }
        ids = append(ids, id)
    }
    rows.Close()
    if err := rows.Err(); err != nil {
        return nil, err
    }

    out := []MovieDetail{}
    for _, id := range ids {
        m, err := s.GetMovie(ctx, id)
        if errors.Is(err, ErrNotFound) {
            continue
        }
        if err != nil {
            return nil, err
        }
        out = append(out, m)
    }
    return out, nil
}