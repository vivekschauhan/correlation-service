package service

import (
	"context"

	"github.com/Axway/agent-sdk/pkg/amplify/agent/correlation"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/structpb"
)

type Resource struct {
	Path     string `yaml:"path"`
	APIID    string `yaml:"api_id"`
	Version  string `yaml:"version"`
	Stage    string `yaml:"stage"`
	ClientID string `yaml:"client_id"`
}

type ResourceMappings struct {
	Resources map[string]Resource `yaml:"mapping"`
}

type CorrelationService interface {
	correlation.CorrelationServiceServer
	AddResourceMapping(path string, resource Resource)
}

type correlationService struct {
	CorrelationService
	resMap *ResourceMappings
	log    *logrus.Logger
}

func NewCorrelationService(log *logrus.Logger, resMap *ResourceMappings) CorrelationService {
	return &correlationService{
		resMap: resMap,
		log:    log,
	}
}

func (s *correlationService) GetResourceContext(ctx context.Context, reqCtx *correlation.TransactionContext) (*correlation.ResourceContext, error) {
	s.log.Info("received request")
	s.log.Infof("%+v", reqCtx)
	resCtx := &correlation.ResourceContext{}
	for name, val := range reqCtx.Metadata {
		_, ok := val.GetKind().(*structpb.Value_ListValue)
		if ok {
			s.log.WithField(name, val.GetListValue().AsSlice()).Info("request context metadata")
		} else {
			_, ok = val.GetKind().(*structpb.Value_StructValue)
			if ok {
				s.log.WithField(name, val.GetStructValue().AsMap()).Info("request context metadata")
			} else {
				s.log.WithField(name, val.AsInterface()).Info("request context metadata")
			}
		}
	}

	if reqCtx.Request != nil {
		path, ok := reqCtx.Request.Headers["x-envoy-original-path"]
		if !ok {
			path = reqCtx.Request.Path
			if reqCtx.Request.OriginalPath != "" {
				path = reqCtx.Request.OriginalPath
			}
		}
		res, ok := s.resMap.Resources[path]
		if ok {
			resCtx = &correlation.ResourceContext{
				ApiId:      res.APIID,
				Version:    res.Version,
				Stage:      res.Stage,
				ConsumerId: res.ClientID,
			}
		}
	}

	return resCtx, nil
}

func (s *correlationService) AddResourceMapping(path string, resource Resource) {
	s.resMap.Resources[path] = resource
}
