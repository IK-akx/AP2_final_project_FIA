package grpc

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/fxrnweh9/product-service/internal/domain"
	pb "github.com/fxrnweh9/product-service/proto/v1"
)

type ProductHandler struct {
	pb.UnimplementedProductServiceServer
	svc domain.ProductService
}

func NewProductHandler(svc domain.ProductService) *ProductHandler {
	return &ProductHandler{svc: svc}
}
func (h *ProductHandler) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.ProductResponse, error) {
	p := &domain.Product{
		Name:         req.Name,
		Description:  req.Description,
		CategoryID:   req.CategoryId,
		Price:        req.Price,
		Stock:        int(req.Stock),
		Manufacturer: req.Manufacturer,
		RequiresRX:   req.RequiresRx,
		ImageURL:     req.ImageUrl,
	}

	res, err := h.svc.CreateProduct(ctx, p)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ProductResponse{
		Product: mapProduct(res),
	}, nil
}

func (h *ProductHandler) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.ProductResponse, error) {
	res, err := h.svc.GetProduct(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, "product not found")
	}

	return &pb.ProductResponse{
		Product: mapProduct(res),
	}, nil
}

func (h *ProductHandler) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	products, total, err := h.svc.ListProducts(ctx, int(req.Limit), int(req.Page))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var pbProducts []*pb.Product
	for _, p := range products {
		pbProducts = append(pbProducts, mapProduct(p))
	}

	return &pb.ListProductsResponse{
		Products: pbProducts,
		Total:    int32(total),
	}, nil
}

func (h *ProductHandler) CheckAvailability(ctx context.Context, req *pb.CheckAvailabilityRequest) (*pb.AvailabilityResponse, error) {
	ok, stock, err := h.svc.CheckAvailability(ctx, req.ProductId, int(req.RequestedQuantity))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.AvailabilityResponse{
		ProductId:    req.ProductId,
		IsAvailable:  ok,
		CurrentStock: int32(stock),
	}, nil
}

func (h *ProductHandler) UpdateStock(ctx context.Context, req *pb.UpdateStockRequest) (*pb.StockResponse, error) {
	old, newStock, err := h.svc.UpdateStock(ctx, req.ProductId, int(req.QuantityChange))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.StockResponse{
		ProductId: req.ProductId,
		OldStock:  int32(old),
		NewStock:  int32(newStock),
	}, nil
}

func mapProduct(p *domain.Product) *pb.Product {
	return &pb.Product{
		Id:           p.ID,
		Name:         p.Name,
		Description:  p.Description,
		CategoryId:   p.CategoryID,
		Price:        p.Price,
		Stock:        int32(p.Stock),
		Manufacturer: p.Manufacturer,
		RequiresRx:   p.RequiresRX,
		ImageUrl:     p.ImageURL,
		CreatedAt:    toProtoTime(p.CreatedAt),
		UpdatedAt:    toProtoTime(p.UpdatedAt),
	}
}

func toProtoTime(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}
