package postgres

import "github.com/lib/pq"

const uniquenessViolation = pq.ErrorCode("23505")
