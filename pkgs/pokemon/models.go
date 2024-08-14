package pokemon

type (
	PokemonsResponse struct {
		Count   int `json:"count"`
		Results []struct {
			Name string `json:"name"`
		} `json:"results"`
	}

	PokemonResponse struct {
		Id              int    `json:"id"`
		HeightDecimeter int    `json:"height"`
		WeightHectogram int    `json:"weight"`
		Name            string `json:"name"`
		Stats           []struct {
			Base int `json:"base_stat"`
			Stat struct {
				Name string `json:"name"`
			} `json:"stat"`
		} `json:"stats"`
		StatsTotal int
		Species    struct {
			Name string `json:"name"`
			Url  string `json:"url"`
		} `json:"species"`
		Types []struct {
			Slot int `json:"slot"`
			Type struct {
				Name string `json:"name"`
			} `json:"type"`
		} `json:"types"`
		Abilities []struct {
			Slot     int  `json:"slot"`
			IsHidden bool `json:"is_hidden"`
			Ability  struct {
				Name string `json:"name"`
			} `json:"ability"`
		} `json:"abilities"`
	}
	PokemonSpeciesResponse struct {
		GenderRate  int `json:"gender_rate"`
		CaptureRate int `json:"capture_rate"`
		Varieties   []struct {
			Pokemon struct {
				Name string `json:"name"`
			} `json:"pokemon"`
		} `json:"varieties"`
		EvolutionChain struct {
			Url string `json:"url"`
		} `json:"evolution_chain"`
	}
	chain struct {
		Species struct {
			Name string `json:"name"`
		} `json:"species"`
		IsBaby    bool    `json:"is_baby"`
		EvolvesTo []chain `json:"evolves_to"`
	}
	EvolutionChainResponse struct {
		Chain chain `json:"chain"`
	}
	PokemonDetail struct {
		Info       PokemonResponse
		Species    PokemonSpeciesResponse
		Evolutions EvolutionChainResponse
	}
)
