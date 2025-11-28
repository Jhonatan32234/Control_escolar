package repository

import (
	"api_concurrencia/src/models"
	"gorm.io/gorm"
)

type CuatrimestreRepository struct {
	DB *gorm.DB
}

func NewCuatrimestreRepository(db *gorm.DB) *CuatrimestreRepository {
	return &CuatrimestreRepository{DB: db}
}

// Create crea un nuevo Cuatrimestre en la BD local.
func (r *CuatrimestreRepository) Create(c *models.Cuatrimestre) error {
	return r.DB.Create(c).Error
}

// GetAll obtiene todos los Cuatrimestres.
func (r *CuatrimestreRepository) GetAll() ([]models.Cuatrimestre, error) {
	var cuatrimestres []models.Cuatrimestre
	err := r.DB.Find(&cuatrimestres).Error
	return cuatrimestres, err
}

// GetByID obtiene un Cuatrimestre por ID local.
func (r *CuatrimestreRepository) GetByID(id uint) (models.Cuatrimestre, error) {
	var cuatrimestre models.Cuatrimestre
    // 👈 La precarga es crucial para obtener el ID_Moodle del padre
    err := r.DB.Preload("ProgramaEstudio").First(&cuatrimestre, id).Error 
    return cuatrimestre, err
}

// Update actualiza un Cuatrimestre.
func (r *CuatrimestreRepository) Update(c *models.Cuatrimestre) error {
	return r.DB.Save(c).Error
}

// Delete elimina un Cuatrimestre de la BD local.
func (r *CuatrimestreRepository) Delete(id uint) error {
	// Nota: Se debe implementar la lógica de Moodle (si ya existe) y la verificación de hijos.
	return r.DB.Delete(&models.Cuatrimestre{}, id).Error
}