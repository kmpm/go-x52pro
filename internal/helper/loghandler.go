// SPDX-FileCopyrightText: 2025 Peter Magnusson <me@kmpm.se>
//
// SPDX-License-Identifier: MPL-2.0

package helper

import (
	"context"
	"log"
	"log/slog"
	"os"
)

type FilteredHandlerOptions struct {
	SlogOpts slog.HandlerOptions
}

type FilteredHandler struct {
	slog.Handler
	l *log.Logger
}

func NewFilteredHandler(opts FilteredHandlerOptions) *FilteredHandler {
	h := &FilteredHandler{
		Handler: slog.NewTextHandler(os.Stdout, &opts.SlogOpts),
		l:       log.Default(),
	}
	return h
}

func (h *FilteredHandler) Handle(ctx context.Context, r slog.Record) error {
	return h.Handler.Handle(ctx, r)
}
