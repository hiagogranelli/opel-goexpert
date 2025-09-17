package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
)

const (
	viaCEPURL  = "https://viacep.com.br/ws/%s/json/"
	weatherURL = "https://api.weatherapi.com/v1/current.json?key=%s&q=%s&aqi=no"
)

var (
	serviceName   = os.Getenv("OTEL_SERVICE_NAME")
	collectorURL  = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	weatherAPIKey string
)

type ViaCEPResponse struct {
	Localidade string `json:"localidade"`
}

type WeatherResponse struct {
	Current struct {
		TempC float64 `json:"temp_c"`
	} `json:"current"`
}

func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func getTemperatura(c *gin.Context, localidade string) (float64, float64, float64, error) {

	tracer := otel.Tracer("service-b-tracer")
	_, span := tracer.Start(c.Request.Context(), "getTemperatura")
	defer span.End()

	resp, err := http.Get(fmt.Sprintf(weatherURL, weatherAPIKey, url.QueryEscape(localidade)))
	if err != nil {
		return 0, 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, 0, 0, fmt.Errorf("error fetching weather data")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, 0, err
	}

	var openWeatherResponse WeatherResponse
	if err := json.Unmarshal(body, &openWeatherResponse); err != nil {
		return 0, 0, 0, err
	}

	tempC := openWeatherResponse.Current.TempC
	tempF := tempC*1.8 + 32
	tempK := tempC + 273.15

	return tempC, tempF, tempK, nil
}

func getLocalidade(c *gin.Context, cep string) (string, error) {

	tracer := otel.Tracer("service-b-tracer")
	_, span := tracer.Start(c.Request.Context(), "getLocalidade")
	defer span.End()

	resp, err := http.Get(fmt.Sprintf(viaCEPURL, cep))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var viaCEPResponse ViaCEPResponse
	if err := json.Unmarshal(body, &viaCEPResponse); err != nil {
		return "", err
	}

	if viaCEPResponse.Localidade == "" {
		return "", fmt.Errorf("localidade não encontrada")
	}

	return viaCEPResponse.Localidade, nil
}

func climaHandler(c *gin.Context) {

	cep := c.Query("cep")
	if len(cep) != 8 || !isNumeric(cep) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "invalid zipcode"})
		return
	}

	localidade, err := getLocalidade(c, cep)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "cannot find zipcode"})
		return
	}

	tempC, tempF, tempK, err := getTemperatura(c, localidade)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "error fetching weather data"})
		return
	}

	response := map[string]float64{
		"temp_C": tempC,
		"temp_F": tempF,
		"temp_K": tempK,
	}

	c.JSON(http.StatusOK, response)
}

func main() {
	godotenv.Load()
	weatherAPIKey = os.Getenv("WEATHER_API_KEY")

	// Initialize OpenTelemetry
	ctx := context.Background()
	_, shutdown, err := InitTracer(ctx, serviceName, collectorURL)

	if err != nil {
		log.Fatalf("failed to initialize OpenTelemetry: %s, %v", collectorURL, err)
	}
	defer shutdown(ctx)

	r := gin.Default()
	r.Use(otelgin.Middleware(serviceName))
	r.GET("/temperatura", climaHandler)
	r.Run(":8081")
}
