# 🚛 Route Manager API

Backend robusto y escalable para la gestión de rutas logísticas y conductores. Este proyecto implementa una arquitectura limpia, seguridad basada en roles (RBAC) y autenticación mediante OAuth de Google.

![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)
![Gin Framework](https://img.shields.io/badge/Gin-Framework-ff5a5f?style=flat&logo=go)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-Supabase-336791?style=flat&logo=postgresql)
![Status](https://img.shields.io/badge/Status-Development-yellow)

## 🚀 Tecnologías

* **Lenguaje:** Golang
* **Web Framework:** Gin Gonic
* **ORM:** GORM (Driver Postgres)
* **Base de Datos & Auth:** Supabase (PostgreSQL + Auth)
* **Seguridad:** JWT (JSON Web Tokens) validado vía JWKS (JSON Web Key Set).

## 🏗 Arquitectura

El proyecto sigue una estructura modular orientada al dominio para facilitar la escalabilidad y el mantenimiento:

```text
├── api
│   ├── config       # Configuración de entorno
│   ├── database     # Conexión a BD (Singleton)
│   ├── domains      # Modelos de datos (Structs)
│   ├── handlers     # Controladores / Lógica de negocio
│   │   ├── auth     # Registro y Sincronización
│   │   ├── health   # Health Checks (Liveness/Readiness)
│   │   ├── routes   # Gestión de Rutas y Waypoints
│   │   └── users    # Gestión de Usuarios (CRUD)
│   └── middleware   # Auth (JWKS) y Roles (RBAC)
└── main.go          # Punto de entrada y definición de rutas

✨ Funcionalidades Principales

🔐 Seguridad y Autenticación
• Integración con Google OAuth: Validación de tokens mediante JWKS de Supabase.
• RBAC (Role-Based Access Control): Middleware personalizado para restringir endpoints a roles específicos (admin, driver).
• Sincronización Automática: Registro de usuarios en BD local al primer login, extrayendo metadatos (avatar, email verificado).

👤 Gestión de Usuarios
• Perfil propio (/me).
• CRUD completo para Administradores (Listar, Ver detalle, Editar, Eliminar - Soft Delete).

📍 Gestión de Rutas

• Creación Transaccional: Creación de rutas con múltiples paradas (waypoints) en una sola petición.
• Asignación de Conductores: Vinculación de rutas a usuarios con rol driver.
• Listado Inteligente: Los administradores ven todo; los conductores solo ven sus rutas asignadas.

💓 DevOps

• Health Check: Endpoint /health que verifica la conectividad HTTP y el estado de la conexión a la base de datos (ideal para Cron Jobs y Uptime Monitors).

🛠️ Instalación y Configuración

• Clonar el repositorio: git clone [https://github.com/GalahadDev/Portfolio-1-back/tree/main]
                                    cd route-manager
• Configurar Variables de Entorno: Crea un archivo .env en la raíz del proyecto basándote en el ejemplo

    PORT=8080
    DATABASE_URL="postgres://usuario:password@host:port/database"
    SUPABASE_URL="[https://tu-proyecto.supabase.co](https://tu-proyecto.supabase.co)"

• Instalar Dependencias: go mod tidy

• Ejecutar el Servidor: go run main.go

📡 Endpoints API

• Health: Estado del servicio y DB

    GET: /api/v1/health: Sincronizar usuario Google

• Auth: Sincronizar usuario Google

    POST: /api/v1/auth/register

• Usuarios: 

    GET	/api/v1/users/me	    Ver mi perfil
    GET	/api/v1/users	        Listar usuarios
    GET	/api/v1/users/:id	    Ver usuario por ID
    PUT	/api/v1/users/:id	    Editar usuario
    DELETE	/api/v1/users/:id	Eliminar usuario

•  Rutas

    GET	/api/v1/routes	                Listar rutas
    POST    /api/v1/routes	            Crear ruta + Waypoints
    PATCH	/api/v1/routes/:id/assign	Asignar conductor
