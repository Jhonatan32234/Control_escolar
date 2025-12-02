package repository

import (
	"api_concurrencia/src/models"
	"gorm.io/gorm"
)

type UsuarioRepository struct {
	DB *gorm.DB
}

func NewUsuarioRepository(db *gorm.DB) *UsuarioRepository {
	return &UsuarioRepository{DB: db}
}

// Create crea un nuevo Usuario en la BD local.
func (r *UsuarioRepository) Create(u *models.Usuario) error {
	return r.DB.Create(u).Error
}

// GetByID obtiene un Usuario por ID local.
func (r *UsuarioRepository) GetByID(id uint) (models.Usuario, error) {
	var u models.Usuario
	err := r.DB.First(&u, id).Error
	return u, err
}

// GetAll obtiene todos los Usuarios.
func (r *UsuarioRepository) GetAll() ([]models.Usuario, error) {
	var usuarios []models.Usuario
	err := r.DB.Find(&usuarios).Error
	return usuarios, err
}

// Update actualiza un Usuario.
func (r *UsuarioRepository) Update(u *models.Usuario) error {
	return r.DB.Save(u).Error
}

// Delete elimina un Usuario de la BD local.
func (r *UsuarioRepository) Delete(id uint) error {
	return r.DB.Delete(&models.Usuario{}, id).Error
}

// GetUnsyncedByRole obtiene usuarios que aún no tienen ID_Moodle, filtrados por rol.
func (r *UsuarioRepository) GetUnsyncedByRole(role string) ([]models.Usuario, error) {
	var usuarios []models.Usuario
	// Buscamos donde ID_Moodle es NULL Y el Rol coincide.
	err := r.DB.Where("id_moodle IS NULL AND rol = ?", role).Find(&usuarios).Error
	return usuarios, err
}


// GetByGroupID obtiene todos los Usuarios que pertenecen a un Grupo.
func (r *UsuarioRepository) GetByGroupID(grupoID uint) ([]models.Usuario, error) {
    var usuarios []models.Usuario
    // Une implícitamente con la tabla de unión 'usuario_grupos'
    err := r.DB.
        Joins("JOIN usuario_grupos ug ON ug.usuario_id = usuarios.id").
        Where("ug.grupo_id = ?", grupoID).
        Find(&usuarios).Error
    return usuarios, err
}