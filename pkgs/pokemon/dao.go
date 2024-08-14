package pokemon

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func GetPokemon(name string) PokemonDetail {
	data := new(PokemonDetail)
	{
		url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s/", name)
		resp, err := http.Get(url)

		if err != nil {
			fmt.Println(err)
		}

		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
		}

		if err := json.Unmarshal(body, &data.Info); err != nil {
			fmt.Println(err)
		}
	}
	{
		url := data.Info.Species.Url
		resp, err := http.Get(url)

		if err != nil {
			fmt.Println(err)
		}

		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
		}

		if err := json.Unmarshal(body, &data.Species); err != nil {
			fmt.Println(err)
		}
	}
	{
		url := data.Species.EvolutionChain.Url
		resp, err := http.Get(url)

		if err != nil {
			fmt.Println(err)
		}

		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
		}

		if err := json.Unmarshal(body, &data.Evolutions); err != nil {
			fmt.Println(err)
		}
	}

	return *data
}

func GetPokemons() PokemonsResponse {
	url := "https://pokeapi.co/api/v2/pokemon/?limit=99999"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	data := new(PokemonsResponse)
	if err := json.Unmarshal(body, &data); err != nil {
		fmt.Println(err)
	}

	return *data
}
