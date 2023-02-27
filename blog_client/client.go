package main

import (
	"context"
	"fmt"
	"log"

	"github.com/mohdjishin/gRPC-blog-service/blogpb"
	"google.golang.org/grpc"
)

func main() {

	fmt.Println("Hello,I'm a client")

	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}
	defer conn.Close()
	c := blogpb.NewBlogServiceClient(conn)

	// blog := &blogpb.Blog{
	// 	AuthorId: "Jishin",
	// 	Title:    "My first blog",
	// 	Content:  "Content of the first blog",
	// }

	// upBlog := &blogpb.Blog{

	// 	Id:       "63fc4b7c2ea1c2ea73c92807",
	// 	Title:    "My first blog (edited v1)",
	// 	Content:  "Content of the first blog, with some awesome additions!",
	// 	AuthorId: "Mohd Jishin Jamal",
	// }

	// doCreateBlog(c, blog)
	// doReadBlog(c, "63fc4b7c2ea1c2ea73c92807")
	// doUpateBlog(c, upBlog)

	doDeleteBlog(c, "63fc4e94adfe3ae273047a94")
}

func doDeleteBlog(c blogpb.BlogServiceClient, id string) {
	res, err := c.DeleteBlog(context.Background(), &blogpb.DeleteBlogRequest{BlogId: id})
	if err != nil {
		fmt.Printf("Error happened while deleting: %v", err)

	}
	fmt.Println("Blog was deleted: ", res)
}

func doUpateBlog(c blogpb.BlogServiceClient, blog *blogpb.Blog) {

	res, err := c.UpdateBlog(context.Background(), &blogpb.UpdateBlogRequest{
		Blog: blog,
	},
	)

	if err != nil {
		log.Fatalf("Unexpected error: %v", err)
	}

	fmt.Printf("Blog has been updated %v", res)

}
func doCreateBlog(c blogpb.BlogServiceClient, blog *blogpb.Blog) {

	fmt.Println("Creating the blog")
	createBlogRes, err := c.CreateBlog(context.Background(), &blogpb.CreateBlogRequest{
		Blog: blog,
	})

	if err != nil {
		log.Fatalf("Unexpected error: %v", err)
	}

	fmt.Printf("Blog has been created %v", createBlogRes)

}

func doReadBlog(c blogpb.BlogServiceClient, id string) {

	res, err := c.ReadBlog(context.Background(), &blogpb.ReadBlogRequest{BlogId: "63fc4b7c2ea1c2ea73c92807"})

	if err != nil {
		fmt.Printf("Error happened while reading: %v", err)

	}
	fmt.Println("Blog was read: ", res)

}
