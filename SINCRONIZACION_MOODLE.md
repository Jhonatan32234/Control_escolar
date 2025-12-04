# Sistema de Sincronización y Actualización con Moodle

## Problema 1: Actualizar datos en Moodle cuando cambias algo local

### ✅ SOLUCIÓN IMPLEMENTADA

Cuando modificas un registro en tu base de datos local (por ejemplo, cambias el nombre de un usuario), ahora puedes actualizar automáticamente ese cambio en Moodle.

### Cómo funciona:

1. **Endpoint de Sync mejorado** - `POST /usuario/sync/{id}`
   - Si el usuario **NO tiene** `ID_Moodle` → Lo **CREA** en Moodle
   - Si el usuario **YA tiene** `ID_Moodle` → Lo **ACTUALIZA** en Moodle

2. **Funciones de UPDATE agregadas en Moodle Client:**
   - `core_user_update_users` - Actualiza usuarios
   - `core_course_update_categories` - Actualiza cuatrimestres (subcategorías)
   - `core_course_update_courses` - Actualiza asignaturas (cursos)

### Ejemplo de uso:

```bash
# 1. Cambias el nombre de un usuario en tu BD local
PUT /usuario/1
{
  "first_name": "Juan Carlos",  # Antes era "Juan"
  "last_name": "Pérez García"
}

# 2. Sincronizas el cambio a Moodle
POST /usuario/sync/1

# El sistema detecta que ya tiene ID_Moodle y ejecuta UPDATE en lugar de CREATE
```

### Qué se actualiza en Moodle:
- **Usuarios**: Username, Firstname, Lastname, Email, IDNumber (Matrícula)
- **Cuatrimestres**: Name, IDNumber, Description
- **Asignaturas**: Fullname, Shortname, Summary, IDNumber

---

## Problema 2: Subir gran cantidad de datos a Moodle

### ✅ SOLUCIÓN IMPLEMENTADA

Sistema de **sincronización masiva** con procesamiento por lotes y concurrencia.

### Endpoint de sincronización masiva:

```bash
POST /usuario/bulk-sync?role=Docente
POST /usuario/bulk-sync?role=Alumno
```

### Cómo funciona:

1. **Búsqueda inteligente**: Encuentra todos los usuarios que NO tienen `ID_Moodle` (nunca sincronizados)
2. **Procesamiento por lotes**: Divide en grupos de 100 usuarios
3. **Concurrencia**: Procesa múltiples lotes en paralelo (usando goroutines)
4. **Actualización automática**: Guarda el `ID_Moodle` de cada usuario en tu BD local

### Ejemplo: Subir 875 alumnos a Moodle

```bash
# 1. Tienes 875 alumnos en tu BD (del archivo bd.txt)
# 2. Llamas al endpoint:
POST /usuario/bulk-sync?role=Alumno
Authorization: Bearer <tu-token-jwt>

# Respuesta inmediata:
{
  "message": "Sincronización masiva de usuarios con rol 'Alumno' iniciada en segundo plano..."
}

# El proceso se ejecuta en background:
# - Lote 1: Usuarios 1-100 → API Moodle
# - Lote 2: Usuarios 101-200 → API Moodle
# - Lote 3: Usuarios 201-300 → API Moodle
# ... (simultáneamente)
# - Lote 9: Usuarios 801-875 → API Moodle

# Revisa los logs del servidor para ver el progreso:
# ✅ Usuario 'alumno001' sincronizado con Moodle ID: 3456
# ✅ Usuario 'alumno002' sincronizado con Moodle ID: 3457
# ...
```

### Ventajas del sistema por lotes:

✅ **Rápido**: Procesa 875 usuarios en ~9 llamadas en lugar de 875 llamadas individuales
✅ **Robusto**: Si un lote falla, los demás continúan
✅ **No bloquea**: El API responde inmediatamente (HTTP 202 Accepted)
✅ **Trazable**: Los logs muestran el progreso en tiempo real

---

## Flujo completo recomendado

### Escenario 1: Cargar docentes nuevos a Moodle

```bash
# 1. Crear docentes en BD local (ya tienes 35 docentes en bd.txt)

# 2. Sincronizar todos los docentes a Moodle de golpe
POST /usuario/bulk-sync?role=Docente
Authorization: Bearer <token>

# 3. Espera unos segundos y verifica en los logs del servidor
# Verás: "✅ Sincronización masiva de usuarios de rol Docente finalizada."
```

### Escenario 2: Corregir datos y actualizar en Moodle

```bash
# 1. Te equivocaste en el email de un docente
GET /usuario/5
# Ves que el docente tiene ID_Moodle: 2345

# 2. Corriges el email local
PUT /usuario/5
{
  "email": "luis.garcia.correcto@universidad.edu.mx"
}

# 3. Actualizas en Moodle
POST /usuario/sync/5

# El sistema detecta que tiene ID_Moodle y ejecuta UPDATE
# Log: "✅ Usuario 'docente03' (Moodle ID: 2345) actualizado exitosamente en Moodle"
```

### Escenario 3: Workflow completo para un nuevo semestre

```bash
# 1. Programa de Estudio
POST /programa-estudio
POST /programa-estudio/sync/1

# 2. Cuatrimestres (10 cuatrimestres)
POST /cuatrimestre (x10 veces con diferentes datos)
# Sincronizar todos individualmente o implementar bulk-sync

# 3. Asignaturas (70 asignaturas)
POST /asignatura (x70 veces)
# Sincronizar todas

# 4. Usuarios (875 alumnos + 35 docentes)
# Ya los tienes en BD (bd.txt)
POST /usuario/bulk-sync?role=Alumno   # Sube 875 alumnos
POST /usuario/bulk-sync?role=Docente  # Sube 35 docentes

# 5. Grupos
POST /grupo (crear grupos)
POST /grupo/sync/{id}

# 6. Matricular alumnos
POST /usuario/enrol/{usuarioID}/{asignaturaID}
```

---

## Endpoints disponibles para sincronización

### Usuarios
- `POST /usuario/sync/{id}` - Sincroniza 1 usuario (CREATE o UPDATE)
- `POST /usuario/bulk-sync?role=<Docente|Alumno>` - Sincroniza todos los no sincronizados

### Cuatrimestres
- `POST /cuatrimestre/sync/{id}` - Sincroniza 1 cuatrimestre

### Asignaturas
- `POST /asignatura/sync/{id}` - Sincroniza 1 asignatura

### Grupos
- `POST /grupo/sync/{id}` - Sincroniza 1 grupo
- `POST /grupo/add-members/{grupoID}` - Agrega miembros al grupo

### Programas de Estudio
- `POST /programa-estudio/sync/{id}` - Sincroniza 1 programa

---

## Monitoreo y logs

El servidor registra cada operación:

```
[INFO] Iniciando sincronización masiva para 875 usuarios de rol Alumno...
[INFO] -> Procesando lote de 100 usuarios...
[INFO] -> Procesando lote de 100 usuarios...
[INFO] DEBUG: 100 usuarios codificados.
[INFO] ✅ Usuario 'alumno001' sincronizado con Moodle ID: 3456
[INFO] ✅ Usuario 'alumno002' sincronizado con Moodle ID: 3457
...
[INFO] ✅ Sincronización masiva de usuarios de rol Alumno finalizada.
```

---

## Configuración en .env

```env
# Moodle
MOODLE_URL=https://tu-moodle.com
MOODLE_TOKEN=tu_token_ws_aqui

# JWT para autenticación
JWT_SECRET=tu-secret-super-seguro

# Base de datos
DATABASE_URL=user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000,https://miapp.com
```

---

## Resumen

✅ **Actualizar datos en Moodle**: Usa `POST /sync/{id}` después de modificar un registro local
✅ **Carga masiva**: Usa `POST /bulk-sync?role=X` para subir cientos de usuarios de golpe
✅ **Procesamiento inteligente**: Sistema detecta automáticamente si debe crear o actualizar
✅ **Sin bloqueos**: Las operaciones masivas corren en background
✅ **Trazabilidad completa**: Todos los logs disponibles para debugging
