package main

import (
	"context"
	pb "github.com/lianyz/product/csi"
	"google.golang.org/grpc"
	"io"
	"log"
	"os"
	"strconv"
	"time"
)

const (
	address = "localhost:50051"
)

func toFloat(value string) float32 {
	f, err := strconv.ParseFloat(value, 32)
	if err != nil {
		return float32(0.0)
	}
	return float32(f)
}

func initItems(args ...string) []string {
	items := make([]string, 0)
	for _, item := range args {
		items = append(items, item)
	}

	return items
}

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewProductInfoClient(conn)

	if len(os.Args) < 4 {
		log.Fatalf("usage: client product_name product_description product_price")
	}
	name := os.Args[1]
	description := os.Args[2]
	price := toFloat(os.Args[3])

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	log.Printf("add product price: %f", price)

	res, err := client.AddProduct(ctx,
		&pb.AddProductRequest{Name: name,
			Description: description,
			Price:       price})
	if err != nil {
		log.Fatalf("Could not add product: %v", err)
	}

	log.Printf("Product ID: %s added successfuly", res.Value)

	product, err := client.GetProduct(ctx, &pb.GetProductRequest{Value: res.Value})
	if err != nil {
		log.Fatalf("Could not get product: %v", err)
	}
	log.Printf("Product: %v", product)
	log.Printf("Product Price: %f", product.Price)

	log.Printf("************************")
	streamClient := pb.NewOrderManagementClient(conn)
	searchStream, _ := streamClient.SearchOrders(ctx, &pb.SearchOrdersRequest{Item: "Google"})

	log.Printf("begin Orders")
	for {
		order, err := searchStream.Recv()
		if err == io.EOF {
			log.Print("EOF")
			break
		}
		log.Print("Search Result: ", order)
	}
	log.Printf("end Orders")

	log.Printf("************************")

	updateStream, err := streamClient.UpdateOrders(ctx)
	if err != nil {
		log.Fatalf("%v.UpdateOrders(_) = +, %v", streamClient, err)
	}

	upOrder1 := &pb.Order{
		Id:          "11",
		Description: "order11",
		Price:       float32(1100.0),
		Destination: "client11",
		Items:       initItems("Google", "Apple", "Baidu"),
	}
	if err := updateStream.Send(upOrder1); err != nil {
		log.Fatalf("update order %v failed.", upOrder1)
	}

	upOrder2 := &pb.Order{
		Id:          "12",
		Description: "order12",
		Price:       float32(1200.0),
		Destination: "client12",
		Items:       initItems("Google", "Apple", "Baidu"),
	}
	if err := updateStream.Send(upOrder2); err != nil {
		log.Fatalf("update order %v failed.", upOrder2)
	}

	upOrder3 := &pb.Order{
		Id:          "13",
		Description: "order13",
		Price:       float32(1300.0),
		Destination: "client13",
		Items:       initItems("Google", "Apple", "Baidu"),
	}
	if err := updateStream.Send(upOrder3); err != nil {
		log.Fatalf("update order %v failed.", upOrder3)
	}

	updateRes, err := updateStream.CloseAndRecv()
	if err != nil {
		log.Fatalf("CloseAndRecv failed %v", err)
	}
	log.Printf("Update Orders Res: %s", updateRes)
}
