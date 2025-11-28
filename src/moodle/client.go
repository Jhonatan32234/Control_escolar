package moodle

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type Client struct {
	BaseURL string
	Token   string
}

func NewClient() *Client {
	// Es crucial usar variables de entorno para estos valores
	return &Client{
		BaseURL: os.Getenv("MOODLE_URL"), 
		Token:   os.Getenv("MOODLE_TOKEN"),
	}
}

// Call ejecuta una llamada genérica al WebService de Moodle.
func (c *Client) Call(function string, data interface{}, response interface{}) error {
	if c.BaseURL == "" || c.Token == "" {
		return fmt.Errorf("URL y Token de Moodle no configurados")
	}

	log.Printf("URL: "+c.BaseURL," Token: "+c.Token)

	// 1. Determinar la clave de la función
    var functionKey string
    switch function {
    case "core_course_create_categories":
        functionKey = "categories"
    case "core_course_create_courses":
        functionKey = "courses"
    case "core_user_create_users":
        functionKey = "users"
    default:
        return fmt.Errorf("función Moodle desconocida: %s. No se puede determinar la clave del payload", function)
    }

	log.Printf("Función Moodle: "+function," Clave de datos: "+functionKey)

	postBody := url.Values{}

	postBody.Set("wstoken", c.Token)
	postBody.Set("wsfunction", function)
	postBody.Set("moodlewsrestformat", "json")

	log.Printf("Preparando datos para función "+function)
    // **NUEVA LÓGICA:** Aplanar la estructura de datos
    

	switch function {
case "core_course_create_categories":
    categories, ok := data.([]CategoryRequest)
    if !ok {
        return fmt.Errorf("error de tipo: se esperaba []CategoryRequest")
    }
    for i, cat := range categories {
        prefix := fmt.Sprintf("%s[%d]", functionKey, i)
        postBody.Set(fmt.Sprintf("%s[name]", prefix), cat.Name)
        postBody.Set(fmt.Sprintf("%s[parent]", prefix), fmt.Sprintf("%d", cat.Parent))
        if cat.IDNumber != "" {
            postBody.Set(fmt.Sprintf("%s[idnumber]", prefix), cat.IDNumber)
        }
        if cat.Description != "" {
            postBody.Set(fmt.Sprintf("%s[description]", prefix), cat.Description)
        }
    }
    log.Printf("DEBUG: %d categorías codificadas.", len(categories))

case "core_course_create_courses":
    courses, ok := data.([]CourseRequest)
    if !ok {
        return fmt.Errorf("error de tipo: se esperaba []CourseRequest")
    }
    for i, course := range courses {
        prefix := fmt.Sprintf("%s[%d]", functionKey, i)
        postBody.Set(fmt.Sprintf("%s[fullname]", prefix), course.Fullname)
        postBody.Set(fmt.Sprintf("%s[shortname]", prefix), course.Shortname)
        postBody.Set(fmt.Sprintf("%s[categoryid]", prefix), fmt.Sprintf("%d", course.Categoryid))
        postBody.Set(fmt.Sprintf("%s[visible]", prefix), "1") // Forzar visibilidad
        if course.IDNumber != "" {
            postBody.Set(fmt.Sprintf("%s[idnumber]", prefix), course.IDNumber)
        }
        if course.Summary != "" {
            postBody.Set(fmt.Sprintf("%s[summary]", prefix), course.Summary)
        }
    }
    log.Printf("DEBUG: %d cursos codificados.", len(courses))

    default:

    }

	
	log.Printf("Datos preparados para función "+function+": "+postBody.Encode())
	urlMoodle := fmt.Sprintf("%s/webservice/rest/server.php", c.BaseURL)

	log.Printf("URL Moodle: "+urlMoodle)
	log.Printf("Body: "+postBody.Encode())
	resp, err := http.Post(
        urlMoodle, 
        "application/x-www-form-urlencoded", 
        strings.NewReader(postBody.Encode()), // Codifica los parámetros como 'key=value&key2=value2'
    )
    if err != nil {
        return fmt.Errorf("error al enviar petición a Moodle: %w", err)
    }
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// 3. Manejo de errores de Moodle o HTTP
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("moodle devolvió un error HTTP %d: %s", resp.StatusCode, string(body))
	}

	// Moodle devuelve un array, o un objeto de error.
	var moodleError struct {
		Errorcode string `json:"errorcode"`
		Message   string `json:"message"`
	}
	if err := json.Unmarshal(body, &moodleError); err == nil && moodleError.Errorcode != "" {
		return fmt.Errorf("error de API de Moodle (%s): %s", moodleError.Errorcode, moodleError.Message)
	}

	// 4. Decodificar la respuesta exitosa
	if err := json.Unmarshal(body, response); err != nil {
		return fmt.Errorf("error al decodificar respuesta de Moodle: %w. Cuerpo: %s", err, string(body))
	}

	return nil
}