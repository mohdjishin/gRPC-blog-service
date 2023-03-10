package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/mohdjishin/gRPC-blog-service/blogpb"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc/reflection"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	blogpb.UnimplementedBlogServiceServer
}

var collection *mongo.Collection

type blogItem struct {
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Title    string             `json:"title"`
	Content  string             `json:"content"`
	AuthorId string             `json:"author_id"`
}

func main() {

	// if we crash the go code, we get the file name and line number
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Println("Blog service started!")

	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	err = client.Connect(context.TODO())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	collection = client.Database("mydb").Collection("blog")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")

	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{}

	s := grpc.NewServer(opts...)
	reflection.Register(s)

	// s := grpc.NewServer()

	blogpb.RegisterBlogServiceServer(s, &server{})

	go func() {

		fmt.Println("Starting server...")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}

	}()

	// Wait for control C to exit
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	// Block until a signal is received
	<-ch
	fmt.Println("Stopping the server")
	s.Stop()
	fmt.Println("Closing the listener")
	lis.Close()
	fmt.Println("Closing MongoDB connection")
	client.Disconnect(context.TODO())
	fmt.Println("End of program")

}

func (s *server) CreateBlog(ctx context.Context, req *blogpb.CreateBlogRequest) (*blogpb.CreateBlogResponse, error) {
	fmt.Println("Create blog request")

	blog := req.GetBlog()

	data := blogItem{
		Title:    blog.GetTitle(),
		Content:  blog.GetContent(),
		AuthorId: blog.GetAuthorId(),
	}

	res, err := collection.InsertOne(context.Background(), data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Internal error %v", err)
	}

	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Errorf(codes.Internal, "Cannot convert to OID")
	}

	return &blogpb.CreateBlogResponse{
		Blog: &blogpb.Blog{
			Id:       oid.Hex(),
			Title:    data.Title,
			Content:  data.Content,
			AuthorId: data.AuthorId,
		},
	}, nil
}

func (s *server) ReadBlog(ctx context.Context, req *blogpb.ReadBlogRequest) (*blogpb.ReadBlogResponse, error) {
	fmt.Println("Read blog request")

	blogId := req.GetBlogId()
	oid, err := primitive.ObjectIDFromHex(blogId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Cannot parse ID")
	}
	data := blogItem{}

	res := collection.FindOne(context.Background(), primitive.M{"_id": oid})

	if err := res.Decode(&data); err != nil {
		return nil, status.Errorf(codes.NotFound, "Cannot find blog with specified ID")
	}

	// fmt.Println(res)
	return &blogpb.ReadBlogResponse{
		Blog: &blogpb.Blog{

			Id:       string(data.ID.Hex()),
			Title:    data.Title,
			Content:  data.Content,
			AuthorId: data.AuthorId,
		},
	}, nil
}

func (s *server) UpdateBlog(ctx context.Context, in *blogpb.UpdateBlogRequest) (*blogpb.UpdateBlogResponse, error) {
	fmt.Println("Update blog request")

	blog := in.GetBlog()

	oid, err := primitive.ObjectIDFromHex(blog.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Cannot parse ID")
	}

	// create emply struct
	data := &blogItem{}
	res := collection.FindOne(context.Background(), primitive.M{"_id": oid})
	if err := res.Decode(data); err != nil {
		return nil, status.Errorf(codes.NotFound, "Cannot find blog with specified ID")
	}

	data.Title = blog.GetTitle()
	data.Content = blog.GetContent()
	data.AuthorId = blog.GetAuthorId()

	updateRes, updateErr := collection.ReplaceOne(context.Background(), primitive.M{"_id": oid}, data)
	if updateErr != nil {
		return nil, status.Errorf(codes.Internal, "Cannot update object in MongoDB: %v", updateErr)

	}

	fmt.Println("Update result: ", updateRes)

	return &blogpb.UpdateBlogResponse{
		Blog: &blogpb.Blog{
			Id:       string(data.ID.Hex()),
			Title:    data.Title,
			Content:  data.Content,
			AuthorId: data.AuthorId,
		},
	}, nil

}

func (s *server) ListBlog(req *blogpb.ListBlogRequest, stream blogpb.BlogService_ListBlogServer) error {
	fmt.Println("List blog request")

	cur, err := collection.Find(context.Background(), primitive.D{{}})

	if err != nil {
		return status.Errorf(codes.Internal, "Unknown internal error!: %v", err)
	}

	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {

		data := &blogItem{}
		err := cur.Decode(data)
		if err != nil {
			return status.Errorf(codes.Internal, "Error while decoding data from MongoDB: %v", err)
		}
		cur.Decode(data)
		stream.Send(&blogpb.ListBlogResponse{
			Blog: &blogpb.Blog{
				Id:       string(data.ID.Hex()),
				Title:    data.Title,
				Content:  data.Content,
				AuthorId: data.AuthorId,
			},
		},
		)

		if cur.Err() != nil {
			return status.Errorf(codes.Internal, "Unknown internal error: %v", err)
		}

	}

	return nil

}

func (s *server) DeleteBlog(ctx context.Context, in *blogpb.DeleteBlogRequest) (*blogpb.DeleteBlogResponse, error) {
	fmt.Println("Delete blog request")

	blogId := in.GetBlogId()
	oid, err := primitive.ObjectIDFromHex(blogId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Cannot parse ID")
	}
	res, err := collection.DeleteOne(context.Background(), primitive.M{"_id": oid})

	if err != nil {
		return nil, status.Errorf(codes.Internal, "Cannot delete object in MongoDB: %v", err)

	}
	fmt.Println("Delete result: ", res)
	return &blogpb.DeleteBlogResponse{
		BlogId: blogId}, nil
}
