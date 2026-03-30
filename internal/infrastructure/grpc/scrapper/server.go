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

type ScrapperServiceServer struct {
	pb.UnimplementedScrapperServiceServer
	scrapperService scrapperinterfaces.ScrapperService
	logger          ports.Logger
}

func NewScrapperServiceServer(logger ports.Logger, scrapperService scrapperinterfaces.ScrapperService) *ScrapperServiceServer {
	return &ScrapperServiceServer{
		logger:          logger,
		scrapperService: scrapperService,
	}
}

func (s *ScrapperServiceServer) AddChat(ctx context.Context, req *pb.ChatRequest) (*empty.Empty, error) {
	if err := s.scrapperService.AddChat(req.GetChatId()); err != nil {
		return nil, toStatusError(err)
	}
	return &empty.Empty{}, nil
}

func (s *ScrapperServiceServer) DeleteChat(ctx context.Context, req *pb.ChatRequest) (*empty.Empty, error) {
	if err := s.scrapperService.DeleteChat(req.GetChatId()); err != nil {
		return nil, toStatusError(err)
	}
	return &empty.Empty{}, nil
}

func (s *ScrapperServiceServer) AddLink(ctx context.Context, req *pb.AddLinkRequest) (*empty.Empty, error) {
	link := models.AddLink{
		ChatID: req.GetChatId(),
		URL:    req.GetUrl(),
		Tags:   req.GetTags(),
	}
	if err := s.scrapperService.AddLink(link); err != nil {
		return nil, toStatusError(err)
	}
	return &empty.Empty{}, nil
}

func (s *ScrapperServiceServer) DeleteLink(ctx context.Context, req *pb.DeleteLinkRequest) (*empty.Empty, error) {
	link := models.DeleteLink{
		ChatID: req.GetChatId(),
		URL:    req.GetUrl(),
	}
	if err := s.scrapperService.DeleteLink(link); err != nil {
		return nil, toStatusError(err)
	}
	return &empty.Empty{}, nil
}

func (s *ScrapperServiceServer) ListLinks(ctx context.Context, req *pb.ListLinksRequest) (*pb.ListLinksResponse, error) {
	links, err := s.scrapperService.GetLinks(req.GetChatId(), req.GetTags())
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
