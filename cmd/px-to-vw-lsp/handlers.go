package main

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"go.lsp.dev/protocol"
	"go.uber.org/zap"
)

var log *zap.Logger

// handler
type Handler struct {
	protocol.Server
	documents map[protocol.DocumentURI][]string
	config    *Config
}

func NewHandler(ctx context.Context, server protocol.Server, logger *zap.Logger) (Handler, context.Context, error) {
	log = logger
	config := loadDefaultConfig()
	return Handler{
		Server:    server,
		documents: make(map[protocol.DocumentURI][]string),
		config:    &config,
	}, ctx, nil
}

func (h Handler) Initialize(ctx context.Context, params *protocol.InitializeParams) (*protocol.InitializeResult, error) {
	log.Sugar().Infof("initialize: %v", params)

	// TODO use workspace folders instead... which would require implementing DidChangeWorkspaceFolders
	*h.config = loadConfig(params.RootURI.Filename())

	return &protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			TextDocumentSync: &protocol.TextDocumentSyncOptions{
				OpenClose: true,
				Change:    protocol.TextDocumentSyncKindFull,
			},
			CompletionProvider: &protocol.CompletionOptions{
				TriggerCharacters: []string{"x"},
			},
		},

		ServerInfo: &protocol.ServerInfo{
			Name:    "px-to-vw-lsp",
			Version: "0.1.0",
		},
	}, nil
}

func (h Handler) DidOpen(ctx context.Context, params *protocol.DidOpenTextDocumentParams) error {
	log.Sugar().Infof("didOpen: %v", params)
	h.documents[params.TextDocument.URI] = splitLines(params.TextDocument.Text)
	return nil
}

func (h Handler) DidChange(ctx context.Context, params *protocol.DidChangeTextDocumentParams) error {
	log.Sugar().Infof("didChange: %v", params)
	if len(params.ContentChanges) > 0 {
		h.documents[params.TextDocument.URI] = splitLines(params.ContentChanges[0].Text)
	}
	return nil
}

func (h Handler) DidClose(ctx context.Context, params *protocol.DidCloseTextDocumentParams) error {
	log.Sugar().Infof("didClose: %v", params)
	// TODO is this necessary?
	delete(h.documents, params.TextDocument.URI)
	return nil
}

func (h Handler) Completion(ctx context.Context, params *protocol.CompletionParams) (*protocol.CompletionList, error) {
	log.Sugar().Infof("completion: %v", params)

	// get content
	uri := params.TextDocument.URI
	line := h.documents[uri][params.Position.Line]
	prefix := line[:params.Position.Character]

	// match and parse
	re := regexp.MustCompile(`(\d+(\.\d+)?)px$`)
	match := re.FindStringSubmatch(prefix)
	if match == nil {
		return &protocol.CompletionList{
			IsIncomplete: false,
			Items:        []protocol.CompletionItem{},
		}, nil
	}

	pxValueStr := match[1]
	pxValue, err := strconv.ParseFloat(pxValueStr, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse px value: %v", err)
	}

	// convert to vw
	vwValue := (pxValue / float64(h.config.ViewportWidth)) * 100
	vwValueStr := strconv.FormatFloat(vwValue, 'f', h.config.UnitPrecision, 64)

	// return completion item
	return &protocol.CompletionList{
		IsIncomplete: false,
		Items: []protocol.CompletionItem{
			{
				Kind:       protocol.CompletionItemKindUnit,
				Label:      vwValueStr + "vw",
				FilterText: pxValueStr + "px",

				TextEdit: &protocol.TextEdit{
					Range: protocol.Range{
						Start: protocol.Position{
							Line:      params.Position.Line,
							Character: params.Position.Character - uint32(len(pxValueStr)) - 2,
						},
						End: protocol.Position{
							Line:      params.Position.Line,
							Character: params.Position.Character,
						},
					},
					NewText: vwValueStr + "vw",
				},
			},
		},
	}, nil
}
