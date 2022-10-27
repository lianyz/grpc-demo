/*
@Time : 2022/10/25 23:01
@Author : lianyz
@Description :
*/

package main

import (
	"context"
	"fmt"
	"github.com/gofrs/uuid"
	pb "github.com/lianyz/product/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"strings"
)

type service struct {
	products map[string]*pb.AddProductRequest
	orders   map[string]*pb.Order
}

func (s *service) AddProduct(ctx context.Context, req *pb.AddProductRequest) (*pb.AddProductResponse, error) {
	log.Printf("add product %v", req)

	id, err := uuid.NewV4()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error while generating Product ID", err)
	}

	req.Id = id.String()
	if s.products == nil {
		s.products = make(map[string]*pb.AddProductRequest)
	}
	s.products[req.Id] = req

	return &pb.AddProductResponse{Value: req.Id}, status.New(codes.OK, "").Err()
}

func (s *service) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	product, exists := s.products[req.Value]
	var result *pb.GetProductResponse
	if exists {
		result = &pb.GetProductResponse{
			Id:          product.Id,
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
		}
		return result, status.New(codes.OK, "").Err()
	}

	return nil, status.Errorf(codes.NotFound, "Product does not exist.", req.Value)
}

func (s *service) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.Order, error) {
	order, exists := s.orders[req.Id]
	if exists {
		return order, nil
	}
	return nil, status.Errorf(codes.NotFound, "Order does not exist.", req.Id)
}

func initItems(args ...string) []string {
	items := make([]string, 0)
	for _, item := range args {
		items = append(items, item)
	}

	return items
}

func (s *service) initOrders() {
	s.orders = make(map[string]*pb.Order)

	s.orders["1"] = &pb.Order{
		Id:          "1",
		Description: "order1",
		Price:       float32(100.0),
		Destination: "client1",
		Items:       initItems("Google", "Apple", "Baidu"),
	}

	s.orders["2"] = &pb.Order{
		Id:          "2",
		Description: "order2",
		Price:       float32(200.0),
		Destination: "client2",
		Items:       initItems("Micro", "Apple", "Baidu"),
	}

	s.orders["3"] = &pb.Order{
		Id:          "3",
		Description: "order3",
		Price:       float32(200.0),
		Destination: "client3",
		Items:       initItems("Google", "Apple", "Yahoo"),
	}
}

func (s *service) SearchOrders(req *pb.SearchOrdersRequest, stream pb.OrderManagement_SearchOrdersServer) error {
	if s.orders == nil {
		s.initOrders()
	}

	for key, order := range s.orders {
		log.Printf(key, order)
		for _, item := range order.Items {
			log.Print(item)
			if strings.Contains(item, req.Item) {
				err := stream.Send(order)
				if err != nil {
					return fmt.Errorf("error sending message to stream: %v", err)
				}
			}
		}
	}
	return nil
}

func (s *service) UpdateOrders(stream pb.OrderManagement_UpdateOrdersServer) error {
	orderStr := "Update Order IDs: "
	for {
		order, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(
				&pb.UpdateOrdersResponse{
					Result: "Orders processed " + orderStr,
				})
		}
		if order != nil {
			s.orders[order.Id] = order
			log.Print("Order ID ", order.Id, ": Updated")
			orderStr += order.Id + ", "
		} else {
			log.Print("Order is null")
		}

	}
	return nil
}
