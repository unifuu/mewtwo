package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"

	pb "grpc/proto"

	"google.golang.org/grpc"
)

type pokemonServer struct {
	pb.UnimplementedPokemonServiceServer
}

// PokeAPI response structures
type PokeAPIResponse struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Height int    `json:"height"`
	Weight int    `json:"weight"`
	Types  []struct {
		Type struct {
			Name string `json:"name"`
		} `json:"type"`
	} `json:"types"`
	Sprites struct {
		FrontDefault string `json:"front_default"`
		Other        struct {
			OfficialArtwork struct {
				FrontDefault string `json:"front_default"`
			} `json:"official-artwork"`
		} `json:"other"`
	} `json:"sprites"`
}

func (s *pokemonServer) GetPokemon(ctx context.Context, req *pb.PokemonRequest) (*pb.PokemonResponse, error) {
	query := strings.TrimSpace(strings.ToLower(req.Query))

	if query == "" {
		return &pb.PokemonResponse{
			Success: false,
			Message: "Please enter a Pokemon name or ID",
		}, nil
	}

	log.Printf("Fetching Pokemon: %s", query)

	// Call PokeAPI
	url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s", query)
	resp, err := http.Get(url)
	if err != nil {
		return &pb.PokemonResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to fetch Pokemon: %v", err),
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return &pb.PokemonResponse{
			Success: false,
			Message: "Pokemon not found. Try a different name or ID (1-1025)",
		}, nil
	}

	if resp.StatusCode != 200 {
		return &pb.PokemonResponse{
			Success: false,
			Message: fmt.Sprintf("API error: status code %d", resp.StatusCode),
		}, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &pb.PokemonResponse{
			Success: false,
			Message: "Failed to read API response",
		}, nil
	}

	var pokeData PokeAPIResponse
	if err := json.Unmarshal(body, &pokeData); err != nil {
		return &pb.PokemonResponse{
			Success: false,
			Message: "Failed to parse Pokemon data",
		}, nil
	}

	// Extract types
	types := make([]string, len(pokeData.Types))
	for i, t := range pokeData.Types {
		types[i] = strings.Title(t.Type.Name)
	}

	// Prefer official artwork, fallback to sprite
	imageURL := pokeData.Sprites.Other.OfficialArtwork.FrontDefault
	if imageURL == "" {
		imageURL = pokeData.Sprites.FrontDefault
	}

	pokemon := &pb.Pokemon{
		Id:       int32(pokeData.ID),
		Name:     strings.Title(pokeData.Name),
		Types:    types,
		ImageUrl: imageURL,
		Height:   int32(pokeData.Height),
		Weight:   int32(pokeData.Weight),
	}

	log.Printf("Successfully fetched: %s (ID: %d)", pokemon.Name, pokemon.Id)

	return &pb.PokemonResponse{
		Success: true,
		Message: "Pokemon found!",
		Pokemon: pokemon,
	}, nil
}

func (s *pokemonServer) SearchPokemon(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	// Optional: implement search functionality
	return &pb.SearchResponse{}, nil
}

func main() {
	port := 50051
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterPokemonServiceServer(grpcServer, &pokemonServer{})

	log.Printf("Pokemon gRPC Server listening on port %d", port)
	log.Printf("Ready to fetch Pokemon data from PokeAPI!")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
