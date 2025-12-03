package repository

import (
	"api_concurrencia/src/models"
	"gorm.io/gorm"
)

type ProgramaEstudioRepository struct {
	DB *gorm.DB
}

func NewProgramaEstudioRepository(db *gorm.DB) *ProgramaEstudioRepository {
	return &ProgramaEstudioRepository{DB: db}
}

// Create crea un nuevo Programa de Estudio en la BD local.
func (r *ProgramaEstudioRepository) Create(pe *models.ProgramaEstudio) error {
	return r.DB.Create(pe).Error
}

// GetAll obtiene todos los Programas de Estudio.
func (r *ProgramaEstudioRepository) GetAll() ([]models.ProgramaEstudio, error) {
    var programas []models.ProgramaEstudio
    err := r.DB.Preload("Cuatrimestres").Find(&programas).Error
    return programas, err
}

// GetByID obtiene un Programa de Estudio por ID local.
func (r *ProgramaEstudioRepository) GetByID(id uint) (models.ProgramaEstudio, error) {
    var pe models.ProgramaEstudio
    err := r.DB.Preload("Cuatrimestres").First(&pe, id).Error
    return pe, err
}

// Update actualiza un Programa de Estudio.
func (r *ProgramaEstudioRepository) Update(pe *models.ProgramaEstudio) error {
	return r.DB.Save(pe).Error
}

// Delete elimina un Programa de Estudio de la BD local.
func (r *ProgramaEstudioRepository) Delete(id uint) error {
	return r.DB.Delete(&models.ProgramaEstudio{}, id).Error
}