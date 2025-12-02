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
    case "enrol_manual_enrol_users": 
    functionKey = "enrolments"
    case "core_group_create_groups":      // 👈 NUEVO
    functionKey = "groups"
    case "core_group_add_group_members": // 👈 NUEVO
    functionKey = "members"
    default:
        return fmt.Errorf("función Moodle desconocida: %s. No se puede determinar la clave del payload", function)
    }

	log.Printf("Función Moodle: "+function," Clave de datos: "+functionKey)

	postBody := url.Values{}

	postBody.Set("wstoken", c.Token)
	postBody.Set("wsfunction", function)
	postBody.Set("moodlewsrestformat", "json")

	log.Printf("Datos preparados para función %s: %s", function, postBody.Encode()) // Usar %s
    urlMoodle := fmt.Sprintf("%s/webservice/rest/server.php", c.BaseURL)
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

    case "core_user_create_users":
    users, ok := data.([]UserRequest)
    if !ok {
        return fmt.Errorf("error de tipo: se esperaba []UserRequest")
    }
    for i, user := range users {
        prefix := fmt.Sprintf("%s[%d]", functionKey, i)
        postBody.Set(fmt.Sprintf("%s[username]", prefix), user.Username)
        postBody.Set(fmt.Sprintf("%s[password]", prefix), user.Password)
        postBody.Set(fmt.Sprintf("%s[firstname]", prefix), user.Firstname)
        postBody.Set(fmt.Sprintf("%s[lastname]", prefix), user.Lastname)
        postBody.Set(fmt.Sprintf("%s[email]", prefix), user.Email)
        if user.IDNumber != "" {
            postBody.Set(fmt.Sprintf("%s[idnumber]", prefix), user.IDNumber)
        }
    }
    log.Printf("DEBUG: %d usuarios codificados.", len(users))

    case "enrol_manual_enrol_users":
    enrolments, ok := data.([]EnrolmentRequest)
    if !ok {
        return fmt.Errorf("error de tipo: se esperaba []EnrolmentRequest")
    }
    // La clave real del payload es siempre 'enrolments[i][...]'.
    for i, enrol := range enrolments {
        prefix := fmt.Sprintf("enrolments[%d]", i) // Usamos la clave literal "enrolments"
        postBody.Set(fmt.Sprintf("%s[roleid]", prefix), fmt.Sprintf("%d", enrol.RoleID))
        postBody.Set(fmt.Sprintf("%s[userid]", prefix), fmt.Sprintf("%d", enrol.UserID))
        postBody.Set(fmt.Sprintf("%s[courseid]", prefix), fmt.Sprintf("%d", enrol.CourseID))
    }
    log.Printf("DEBUG: %d matrículas codificadas.", len(enrolments))

    case "core_group_create_groups":
    groups, ok := data.([]GroupRequest)
    if !ok {
        return fmt.Errorf("error de tipo: se esperaba []GroupRequest")
    }
    for i, group := range groups {
        prefix := fmt.Sprintf("%s[%d]", functionKey, i) 
        postBody.Set(fmt.Sprintf("%s[courseid]", prefix), fmt.Sprintf("%d", group.CourseID))
        postBody.Set(fmt.Sprintf("%s[name]", prefix), group.Name)
        postBody.Set(fmt.Sprintf("%s[description]", prefix), group.Description)
        // 💡 NUEVOS CAMPOS OBLIGATORIOS/DEFAULT
        postBody.Set(fmt.Sprintf("%s[descriptionformat]", prefix), fmt.Sprintf("%d", group.DescriptionFormat))
        postBody.Set(fmt.Sprintf("%s[visibility]", prefix), fmt.Sprintf("%d", group.Visibility))
        postBody.Set(fmt.Sprintf("%s[participation]", prefix), fmt.Sprintf("%d", group.Participation))
        
        // Campos Opcionales
        if group.Description != "" {
            postBody.Set(fmt.Sprintf("%s[description]", prefix), group.Description)
        }
        if group.EnrolmentKey != "" {
            postBody.Set(fmt.Sprintf("%s[enrolmentkey]", prefix), group.EnrolmentKey)
        }
        if group.IDNumber != "" {
            postBody.Set(fmt.Sprintf("%s[idnumber]", prefix), group.IDNumber)
        }
    }
    log.Printf("DEBUG: %d grupos codificados.", len(groups))

    case "core_group_add_group_members":
    members, ok := data.([]GroupMemberRequest)
    if !ok {
        return fmt.Errorf("error de tipo: se esperaba []GroupMemberRequest")
    }
    for i, member := range members {
        prefix := fmt.Sprintf("%s[%d]", functionKey, i) // functionKey es "members"
        postBody.Set(fmt.Sprintf("%s[groupid]", prefix), fmt.Sprintf("%d", member.GroupID))
        postBody.Set(fmt.Sprintf("%s[userid]", prefix), fmt.Sprintf("%d", member.UserID))
    }
    log.Printf("DEBUG: %d miembros de grupo codificados.", len(members))
    
    default:

    }

	log.Printf("URL Moodle: "+urlMoodle)
    log.Printf("Body: %s", postBody.Encode()) 
    resp, err := http.Post(
        urlMoodle, 
        "application/x-www-form-urlencoded", 
        strings.NewReader(postBody.Encode()), // Codifica los parámetros como 'key=value&key2=value2'
    )
    if err != nil {
        return fmt.Errorf("error al enviar petición a Moodle: %w", err)
    }
    log.Printf("Respuesta HTTP de Moodle: %s", resp.Status)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// 3. Manejo de errores de Moodle o HTTP
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("moodle devolvió un error HTTP %d: %s", resp.StatusCode, string(body))
	}

    log.Printf("Cuerpo de respuesta de Moodle: %s", string(body))
	// Moodle devuelve un array, o un objeto de error.
	var moodleError struct {
		Errorcode string `json:"errorcode"`
		Message   string `json:"message"`
        Exception string `json:"exception"`
	}

    isError := len(body) > 0 && body[0] == '{'
    if isError {
    if err := json.Unmarshal(body, &moodleError); err == nil && moodleError.Errorcode != "" {
        // Si la decodificación tuvo éxito y Moodle devolvió un error de API
        log.Printf("DEBUG ERROR CHECK Moodle: EXCEPTION: [%s], Código: [%s], Mensaje: [%s]", moodleError.Exception, moodleError.Errorcode, moodleError.Message)
        return fmt.Errorf("error de API de Moodle (%s / %s): %s", moodleError.Exception, moodleError.Errorcode, moodleError.Message)
    } else if err != nil {
        log.Printf("ADVERTENCIA: Falló la decodificación JSON del cuerpo: %v. Cuerpo recibido: %s", err, string(body))
    }
    }
    if err := json.Unmarshal(body, &moodleError); err == nil && moodleError.Errorcode != "" {
    // Si la decodificación tiene éxito Y Moodle devuelve un error de API
    log.Printf("DEBUG ERROR CHECK Moodle: EXCEPTION: [%s], Código: [%s], Mensaje: [%s]", moodleError.Exception, moodleError.Errorcode, moodleError.Message)
    return fmt.Errorf("error de API de Moodle (%s / %s): %s", moodleError.Exception, moodleError.Errorcode, moodleError.Message)
} else if err != nil {
    // Si la decodificación JSON falla (p.ej., el cuerpo es JSON inválido)
    log.Printf("ADVERTENCIA: Falló la decodificación JSON del cuerpo: %v. Cuerpo recibido: %s", err, string(body))
}
    log.Printf("No se detectaron errores en la respuesta de Moodle.")
	// 4. Decodificar la respuesta exitosa
	if err := json.Unmarshal(body, response); err != nil {
		return fmt.Errorf("error al decodificar respuesta de Moodle: %w. Cuerpo: %s", err, string(body))
	}

	return nil
}