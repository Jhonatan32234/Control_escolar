package repository

import (
	"api_concurrencia/src/models"
	"gorm.io/gorm"
)

type AsignaturaRepository struct {
	DB *gorm.DB
}

func NewAsignaturaRepository(db *gorm.DB) *AsignaturaRepository {
	return &AsignaturaRepository{DB: db}
}

// Create crea una nueva Asignatura en la BD local.
func (r *AsignaturaRepository) Create(a *models.Asignatura) error {
	return r.DB.Create(a).Error
}

// GetAll obtiene todas las Asignaturas.
func (r *AsignaturaRepository) GetAll() ([]models.Asignatura, error) {
	var asignaturas []models.Asignatura
	// Preload carga la relación con el Cuatrimestre si la hubiéramos definido
	err := r.DB.Find(&asignaturas).Error
	return asignaturas, err
}

// GetByID obtiene una Asignatura por ID local.
func (r *AsignaturaRepository) GetByID(id uint) (models.Asignatura, error) {
	var asignatura models.Asignatura
	err := r.DB.Preload("Cuatrimestre.ProgramaEstudio").First(&asignatura, id).Error 
	return asignatura, err
}

// Update actualiza una Asignatura.
func (r *AsignaturaRepository) Update(a *models.Asignatura) error {
	return r.DB.Save(a).Error
}

// Delete elimina una Asignatura de la BD local.
func (r *AsignaturaRepository) Delete(id uint) error {
	// Nota: Un curso no debe borrarse si ya tiene usuarios matriculados.
	return r.DB.Delete(&models.Asignatura{}, id).Error
}