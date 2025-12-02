package repository

import (
    "api_concurrencia/src/models"
    "gorm.io/gorm"
)

type GrupoRepository struct {
    DB *gorm.DB
}

func NewGrupoRepository(db *gorm.DB) *GrupoRepository {
    return &GrupoRepository{DB: db}
}

// Create crea un nuevo Grupo en la BD local.
func (r *GrupoRepository) Create(g *models.Grupo) error {
    return r.DB.Create(g).Error
}

// GetByID obtiene un Grupo por ID local.
func (r *GrupoRepository) GetByID(id uint) (models.Grupo, error) {
    var g models.Grupo
    // Preload opcionalmente puedes cargar los usuarios
    err := r.DB.First(&g, id).Error
    return g, err
}

func (r *GrupoRepository) GetAll() ([]models.Grupo, error) {
	var grupos []models.Grupo
	err := r.DB.Find(&grupos).Error
	return grupos, err
}

// Update actualiza un Cuatrimestre.
func (r *GrupoRepository) Update(c *models.Grupo) error {
	return r.DB.Save(c).Error
}

// Delete elimina un Cuatrimestre de la BD local.
func (r *GrupoRepository) Delete(id uint) error {
	return r.DB.Delete(&models.Grupo{}, id).Error
}


// AddMembers añade usuarios a un grupo existente (Actualiza la tabla de unión).
func (r *GrupoRepository) AddMembers(grupoID uint, usuarioIDs []uint) error {
    var grupo models.Grupo
    if err := r.DB.First(&grupo, grupoID).Error; err != nil {
        return err
    }

    var usuarios []models.Usuario
    // Buscar los usuarios por sus IDs
    if err := r.DB.Find(&usuarios, usuarioIDs).Error; err != nil {
        return err
    }

    // GORM maneja la tabla de unión 'usuario_grupos' automáticamente
    return r.DB.Model(&grupo).Association("Usuarios").Append(usuarios)
}

// GetMembers obtiene todos los usuarios de un grupo, incluyendo sus IDs de Moodle.
func (r *GrupoRepository) GetMembers(grupoID uint) ([]models.Usuario, error) {
    var grupo models.Grupo
    // Preload la relación Usuarios.
    err := r.DB.Preload("Usuarios").First(&grupo, grupoID).Error
    if err != nil {
        return nil, err
    }
    return grupo.Usuarios, nil
}