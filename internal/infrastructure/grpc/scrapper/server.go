package scrapper

import (
	"context"
	"errors"
	"sort"

	empty "github.com/golang/protobuf/ptypes/empty"
	scrapperinterfaces "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/interfaces"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/models"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
	ports "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
	pb "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/proto/gen"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ServiceServer struct {
	pb.UnimplementedScrapperServiceServer
	scrapperService scrapperinterfaces.ScrapperService
	logger          ports.Logger
}

func NewScrapperServiceServer(logger ports.Logger, scrapperService scrapperinterfaces.ScrapperService) *ServiceServer {
	return &ServiceServer{
		logger:          logger,
		scrapperService: scrapperService,
	}
}

func (s *ServiceServer) AddChat(ctx context.Context, req *pb.ChatRequest) (*empty.Empty, error) {
	if err := s.scrapperService.AddChat(ctx, req.GetChatId()); err != nil {
		return nil, toStatusError(err)
	}
	return &empty.Empty{}, nil
}

func (s *ServiceServer) DeleteChat(ctx context.Context, req *pb.ChatRequest) (*empty.Empty, error) {
	if err := s.scrapperService.DeleteChat(ctx, req.GetChatId()); err != nil {
		return nil, toStatusError(err)
	}
	return &empty.Empty{}, nil
}

func (s *ServiceServer) AddLink(ctx context.Context, req *pb.AddLinkRequest) (*empty.Empty, error) {
	link := models.AddLink{
		ChatID: req.GetChatId(),
		URL:    req.GetUrl(),
		Tags:   req.GetTags(),
	}
	if err := s.scrapperService.AddLink(ctx, link); err != nil {
		return nil, toStatusError(err)
	}
	return &empty.Empty{}, nil
}

func (s *ServiceServer) DeleteLink(ctx context.Context, req *pb.DeleteLinkRequest) (*empty.Empty, error) {
	link := models.DeleteLink{
		ChatID: req.GetChatId(),
		URL:    req.GetUrl(),
	}
	if err := s.scrapperService.DeleteLink(ctx, link); err != nil {
		return nil, toStatusError(err)
	}
	return &empty.Empty{}, nil
}

func (s *ServiceServer) AddTag(ctx context.Context, req *pb.AddTagRequest) (*empty.Empty, error) {
	tag := models.AddTag{
		ChatID: req.GetChatId(),
		URL:    req.GetUrl(),
		Tag:    req.GetTag(),
	}
	if err := s.scrapperService.AddTag(ctx, tag); err != nil {
		return nil, toStatusError(err)
	}
	return &empty.Empty{}, nil
}

func (s *ServiceServer) DeleteTag(ctx context.Context, req *pb.DeleteTagRequest) (*empty.Empty, error) {
	tag := models.DeleteTag{
		ChatID: req.GetChatId(),
		URL:    req.GetUrl(),
		Tag:    req.GetTag(),
	}
	if err := s.scrapperService.DeleteTag(ctx, tag); err != nil {
		return nil, toStatusError(err)
	}
	return &empty.Empty{}, nil
}

func (s *ServiceServer) GetTags(ctx context.Context, req *pb.GetTagsRequest) (*pb.GetTagsResponse, error) {
	tags, err := s.scrapperService.GetTags(ctx, req.GetChatId(), req.GetUrl())
	if err != nil {
		return nil, toStatusError(err)
	}

	return &pb.GetTagsResponse{
		Tags: tags,
		Size: int32(len(tags)),
	}, nil
}

func (s *ServiceServer) ListLinks(ctx context.Context, req *pb.ListLinksRequest) (*pb.ListLinksResponse, error) {
	links, err := s.scrapperService.GetLinks(ctx, req.GetChatId(), req.GetTags())
	if err != nil {
		return nil, toStatusError(err)
	}

	resp := &pb.ListLinksResponse{
		Links: make([]*pb.Link, 0, len(links)),
		Size:  int32(len(links)),
	}
	for _, link := range links {
		resp.Links = append(resp.Links, toProtoLink(link))
	}

	return resp, nil
}

//nolint:wrapcheck // gRPC handlers must return raw status errors to preserve status codes
func toStatusError(err error) error {
	switch {
	case errors.Is(err, ports.ErrChatNotFound):
		return status.Error(codes.NotFound, ports.ErrChatNotFound.Error())
	case errors.Is(err, ports.ErrLinkNotFound):
		return status.Error(codes.NotFound, ports.ErrLinkNotFound.Error())
	case errors.Is(err, ports.ErrChatAlreadyExists):
		return status.Error(codes.AlreadyExists, ports.ErrChatAlreadyExists.Error())
	case errors.Is(err, ports.ErrLinkAlreadyExists):
		return status.Error(codes.AlreadyExists, ports.ErrLinkAlreadyExists.Error())
	case errors.Is(err, ports.ErrTagAlreadyExists):
		return status.Error(codes.AlreadyExists, ports.ErrTagAlreadyExists.Error())
	case errors.Is(err, ports.ErrTagNotFound):
		return status.Error(codes.NotFound, ports.ErrTagNotFound.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func toProtoLink(link domain.Link) *pb.Link {
	tags := make([]string, 0, len(link.Tags))
	for tag := range link.Tags {
		tags = append(tags, tag)
	}
	sort.Strings(tags)

	return &pb.Link{
		Url:        link.URL,
		Tags:       tags,
		LastUpdate: timestamppb.New(link.LastUpdate),
	}
}
