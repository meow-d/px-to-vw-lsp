package main

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"go.lsp.dev/protocol"
	"go.uber.org/zap"
)

var log *zap.Logger

type Handler struct {
	protocol.Server
	documents        map[protocol.DocumentURI][]string
	workspaceFolders []protocol.WorkspaceFolder
	configs          map[string]*Config
}

func NewHandler(ctx context.Context, server protocol.Server, logger *zap.Logger) (*Handler, context.Context, error) {
	log = logger
	return &Handler{
		Server:    server,
		documents: make(map[protocol.DocumentURI][]string),
		configs:   make(map[string]*Config),
	}, ctx, nil
}

func (h *Handler) Initialize(ctx context.Context, params *protocol.InitializeParams) (*protocol.InitializeResult, error) {
	log.Sugar().Infof("initialize: rootUri=%s, workspaceFolders=%d",
		params.RootURI, len(params.WorkspaceFolders))

	if params.WorkspaceFolders != nil && len(params.WorkspaceFolders) > 0 {
		h.workspaceFolders = params.WorkspaceFolders
		for _, folder := range params.WorkspaceFolders {
			folderPath := strings.TrimPrefix(string(folder.URI), "file://")
			config := loadConfig(folderPath)
			h.configs[folderPath] = &config
			log.Sugar().Infof("Loaded config for workspace folder: %s (viewport: %.0f, precision: %d)",
				folderPath, config.ViewportWidth, config.UnitPrecision)
		}
	} else if params.RootURI != "" {
		rootPath := strings.TrimPrefix(string(params.RootURI), "file://")
		config := loadConfig(rootPath)
		h.configs[rootPath] = &config
		log.Sugar().Warnf("Using deprecated RootURI parameter for initialization")
	}

	supported := true
	return &protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			TextDocumentSync: &protocol.TextDocumentSyncOptions{
				OpenClose: true,
				Change:    protocol.TextDocumentSyncKindFull,
			},
			CompletionProvider: &protocol.CompletionOptions{
				TriggerCharacters: []string{"x"},
			},
			Workspace: &protocol.ServerCapabilitiesWorkspace{
				WorkspaceFolders: &protocol.ServerCapabilitiesWorkspaceFolders{
					Supported:           supported,
					ChangeNotifications: "workspace/didChangeWorkspaceFolders",
				},
			},
		},
		ServerInfo: &protocol.ServerInfo{
			Name:    "px-to-vw-lsp",
			Version: "0.1.0",
		},
	}, nil
}

func (h *Handler) DidChangeWorkspaceFolders(ctx context.Context, params *protocol.DidChangeWorkspaceFoldersParams) error {
	log.Sugar().Infof("didChangeWorkspaceFolders: %v", params)

	for _, removed := range params.Event.Removed {
		removedPath := strings.TrimPrefix(string(removed.URI), "file://")
		delete(h.configs, removedPath)
		for i, folder := range h.workspaceFolders {
			if folder.URI == removed.URI {
				h.workspaceFolders = append(h.workspaceFolders[:i], h.workspaceFolders[i+1:]...)
				break
			}
		}
	}

	for _, added := range params.Event.Added {
		h.workspaceFolders = append(h.workspaceFolders, added)
		addedPath := strings.TrimPrefix(string(added.URI), "file://")
		config := loadConfig(addedPath)
		h.configs[addedPath] = &config
	}

	return nil
}

func (h *Handler) getConfigForDocument(uri protocol.DocumentURI) *Config {
	docPath := strings.TrimPrefix(string(uri), "file://")

	for folderPath, config := range h.configs {
		if strings.HasPrefix(docPath, folderPath) {
			return config
		}
	}

	defaultConfig := loadDefaultConfig()
	return &defaultConfig
}

func (h *Handler) DidOpen(ctx context.Context, params *protocol.DidOpenTextDocumentParams) error {
	uri := params.TextDocument.URI
	lineCount := len(strings.Split(params.TextDocument.Text, "\n"))

	h.documents[uri] = strings.Split(params.TextDocument.Text, "\n")
	log.Sugar().Infof("Document opened: %s (%d lines, %d bytes)",
		uri, lineCount, len(params.TextDocument.Text))
	return nil
}

func (h *Handler) DidChange(ctx context.Context, params *protocol.DidChangeTextDocumentParams) error {
	uri := params.TextDocument.URI

	if len(params.ContentChanges) > 0 {
		h.documents[uri] = strings.Split(params.ContentChanges[0].Text, "\n")
		log.Sugar().Debugf("Document changed: %s (%d content changes)",
			uri, len(params.ContentChanges))
	}
	return nil
}

func (h *Handler) DidClose(ctx context.Context, params *protocol.DidCloseTextDocumentParams) error {
	uri := params.TextDocument.URI

	// Clean up document tracking when file is closed
	delete(h.documents, uri)
	log.Sugar().Debugf("Document closed and cleaned up: %s", uri)

	return nil
}

func (h *Handler) Completion(ctx context.Context, params *protocol.CompletionParams) (*protocol.CompletionList, error) {
	uri := params.TextDocument.URI
	line := h.documents[uri][params.Position.Line]
	prefix := line[:params.Position.Character]

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
		log.Sugar().Warnf("Failed to parse px value '%s' in %s:%d: %v",
			pxValueStr, uri, params.Position.Line, err)
		return nil, fmt.Errorf("failed to parse px value: %v", err)
	}

	config := h.getConfigForDocument(uri)
	vwValue := (pxValue / float64(config.ViewportWidth)) * 100
	vwValueStr := strconv.FormatFloat(vwValue, 'f', config.UnitPrecision, 64)

	log.Sugar().Debugf("Conversion completed: %spx → %svw (viewport: %.0f)",
		pxValueStr, vwValueStr, config.ViewportWidth)

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
