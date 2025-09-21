package ctpgxp

import (
	"context"

	pgx5 "github.com/jackc/pgx/v5"
	"github.com/surkovvs/ct/ctifaces"
)

type (
	tracer struct {
		ctifaces.Logger
	}
)

// Queries

func (tracer tracer) TraceQueryStart(ctx context.Context,
	conn *pgx5.Conn, sd pgx5.TraceQueryStartData,
) context.Context {
	tracer.Debug(
		"query_start",
		"local_addr", conn.PgConn().Conn().LocalAddr().String(),
		"sql", sd.SQL,
		"args", sd.Args)

	return ctx
}

func (tracer tracer) TraceQueryEnd(_ context.Context, conn *pgx5.Conn, ed pgx5.TraceQueryEndData) {
	tracer.Debug(
		"query_end",
		"local_addr", conn.PgConn().Conn().LocalAddr().String(),
		"command_tag", ed.CommandTag.String(),
		"error", ed.Err)
}

// Batch

func (tracer tracer) TraceBatchStart(ctx context.Context,
	conn *pgx5.Conn, sd pgx5.TraceBatchStartData,
) context.Context {
	tracer.Debug(
		"batch_query_start",
		"local_addr", conn.PgConn().Conn().LocalAddr().String(),
		"batch_len", sd.Batch.Len(),
		"queries", sd.Batch.QueuedQueries)
	return ctx
}

func (tracer tracer) TraceBatchQuery(_ context.Context, conn *pgx5.Conn, d pgx5.TraceBatchQueryData) {
	tracer.Debug(
		"batch_query",
		"local_addr", conn.PgConn().Conn().LocalAddr().String(),
		"sql", d.SQL,
		"args", d.Args,
		"error", d.Err)
}

func (tracer tracer) TraceBatchEnd(_ context.Context, conn *pgx5.Conn, ed pgx5.TraceBatchEndData) {
	tracer.Debug(
		"batch_query_end",
		"local_addr", conn.PgConn().Conn().LocalAddr().String(),
		"error", ed.Err)
}

// CopyFrom

func (tracer tracer) TraceCopyFromStart(ctx context.Context,
	conn *pgx5.Conn, sd pgx5.TraceCopyFromStartData,
) context.Context {
	tracer.Debug(
		"copy_from_start",
		"local_addr", conn.PgConn().Conn().LocalAddr().String(),
		"table", sd.TableName.Sanitize(),
		"columns", sd.ColumnNames)
	return ctx
}

func (tracer tracer) TraceCopyFromEnd(_ context.Context, conn *pgx5.Conn, ed pgx5.TraceCopyFromEndData) {
	tracer.Debug(
		"copy_from_end",
		"local_addr", conn.PgConn().Conn().LocalAddr().String(),
		"command_tag", ed.CommandTag.String(),
		"error", ed.Err)
}

// Preparation

func (tracer tracer) TracePrepareStart(ctx context.Context,
	conn *pgx5.Conn, sd pgx5.TracePrepareStartData,
) context.Context {
	tracer.Debug(
		"prepare_start",
		"local_addr", conn.PgConn().Conn().LocalAddr().String(),
		"name", sd.Name,
		"sql", sd.SQL)
	return ctx
}

func (tracer tracer) TracePrepareEnd(_ context.Context, conn *pgx5.Conn, ed pgx5.TracePrepareEndData) {
	tracer.Debug(
		"prepare_end",
		"local_addr", conn.PgConn().Conn().LocalAddr().String(),
		"error", ed.Err)
}

// Connection

func (tracer tracer) TraceConnectStart(ctx context.Context, data pgx5.TraceConnectStartData) context.Context {
	tracer.Info(
		"conn_start",
		"conn_str", data.ConnConfig.ConnString())
	return ctx
}

func (tracer tracer) TraceConnectEnd(_ context.Context, data pgx5.TraceConnectEndData) {
	tracer.Info(
		"conn_end",
		"local_addr", data.Conn.PgConn().Conn().LocalAddr().String(),
		"conn_str", data.Conn.Config().ConnString())
}
