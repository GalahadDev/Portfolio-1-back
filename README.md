# 🚛 Route Manager API (SaaS Multi-tenant)

Backend robusto y escalable para la gestión logística de flotas. Este proyecto implementa una arquitectura **SaaS Multi-tenant** jerárquica, seguridad basada en roles (RBAC), autenticación vía OAuth y un sistema completo de pruebas de entrega (POD).

![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)
![Gin Framework](https://img.shields.io/badge/Gin-Framework-ff5a5f?style=flat&logo=go)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-Supabase-336791?style=flat&logo=postgresql)
![Architecture](https://img.shields.io/badge/Architecture-Clean-green)

## 🚀 Tecnologías

* **Lenguaje:** Golang
* **Web Framework:** Gin Gonic
* **ORM:** GORM (Driver Postgres)
* **Base de Datos & Auth:** Supabase (PostgreSQL + Auth)
* **Storage:** Supabase Storage (Buckets Privados para POD)
* **Seguridad:** JWT validado vía JWKS y Middleware de control de estado.

## 🏗 Arquitectura

El proyecto sigue una estructura modular orientada al dominio (DDD) para facilitar la escalabilidad:

```text
├── api
│   ├── config       # Configuración de entorno
│   ├── database     # Conexión a BD (Singleton)
│   ├── domains      # Modelos de datos (Structs)
│   ├── handlers     # Lógica de negocio
│   │   ├── auth     # Registro
│   │   ├── health   # Health Checks
│   │   ├── routes   # Gestión de Rutas
│   │   ├── users    # Gestión de Usuarios y Flotas
│   │   └── waypoints # Puntos de Entrega & POD
│   ├── middleware   # RBAC, Auth y Validación de Estado
│   ├── services     # Servicios externos (Storage)
│   └── utils        # Generadores de códigos, helpers
└── main.go          # Punto de entrada
```

✨ Funcionalidades Principales

🏢 Gestión de Flotas (Multi-tenancy)

Arquitectura diseñada para aislar la data entre diferentes organizaciones.

    •  Jerarquía de Roles:
            👑 Super Admin: Dueño de la plataforma. Gestiona a los administradores.

            👨‍💼 Admin (Jefe de Flota): Gestiona su propia flota aislada. Solo visualiza sus propios conductores y rutas.

            🚚 Driver: Solo tiene acceso a las rutas que se le han asignado.

    •   Códigos de Flota: Sistema de invitación automática. Los Admins poseen un código único (ej: A7X-99) para enrolar conductores a su equipo.

🔐 Seguridad y Autenticación

    • Seguridad por Defecto: Todo usuario nuevo inicia con estado inactive para prevenir accesos no autorizados.

    • Flujo de Activación: Los conductores se activan automáticamente al consumir el endpoint de Join Fleet con un código válido.
    
    • Auth: Validación de identidad mediante Google OAuth y verificación de tokens (JWKS).

👤 Gestión de Usuarios

    • Perfil propio (/me).

    • CRUD completo para Administradores (Listar, Ver detalle, Editar, Eliminar - Soft Delete).

📍 Gestión de Rutas

    • CRUD Transaccional: Creación de rutas complejas con múltiples paradas (waypoints) asegurando integridad de datos (Atomic Operations).

    • Visibilidad Inteligente: Aislamiento estricto de datos. Un Admin no puede acceder a recursos de otra flota.

    • Proof of Delivery (POD): Subida de evidencia fotográfica a buckets privados en Supabase.
        Generación de Signed URLs temporales para una visualización segura sin exponer el bucket público.

💓 DevOps

• Health Check: Endpoint /health que verifica la conectividad HTTP y el estado de la conexión a la base de datos (ideal para Cron Jobs y Uptime Monitors).

🛠️ Instalación y Configuración

• Clonar el repositorio: git clone [https://github.com/GalahadDev/Portfolio-1-back/tree/main]
                                    cd route-manager
• Configurar Variables de Entorno: Crea un archivo .env en la raíz del proyecto basándote en el ejemplo

    PORT=""
    SUPABASE_URL=""
    SUPABASE_SERVICE_ROLE_KEY=
    DB_USER="" 
    DB_PASSWORD=""
    DB_HOST=""
    DB_PORT=""
    DB_NAME=""

• Instalar Dependencias: go mod tidy

• Ejecutar el Servidor: go run main.go

## 📡 Endpoints API

### 🩺 Health & Auth

| Método | Endpoint | Descripción | Nivel de Acceso |
| --- | --- | --- | --- |
| `GET` | `/api/v1/health` | Verificar estado del servidor y BD | 🟢 Público |
| `POST` | `/api/v1/auth/register` | Registro o Login con Google | 🟢 Público (Con Token) |

### 👥 Usuarios y Gestión de Flotas

| Método | Endpoint | Descripción | Nivel de Acceso |
| --- | --- | --- | --- |
| `POST` | `/api/v1/users/join-fleet` | Unirse a una flota mediante código | 🟡 Usuario Inactivo |
| `GET` | `/api/v1/users/me` | Obtener mi perfil y estado | 🔵 Usuario Activo |
| `GET` | `/api/v1/users` | Listar conductores de mi flota | 🔴 Admin / Super Admin |
| `PUT` | `/api/v1/users/:id` | Promover a Admin o activar usuario | 🔴 Admin / Super Admin |
| `DELETE` | `/api/v1/users/:id` | Eliminar usuario (Soft Delete) | 🔴 Admin / Super Admin |

### 🚚 Rutas (Routes)

| Método | Endpoint | Descripción | Nivel de Acceso |
| --- | --- | --- | --- |
| `GET` | `/api/v1/routes` | Listar rutas (Admin ve todas, Driver ve suyas) | 🔵 Admin / Driver |
| `GET` | `/api/v1/routes/:id` | Ver detalle ruta + **URLs Firmadas** | 🔵 Admin / Driver |
| `POST` | `/api/v1/routes` | Crear nueva ruta con paradas | 🔴 Admin / Super Admin |
| `PUT` | `/api/v1/routes/:id` | Editar datos base de la ruta | 🔴 Admin / Super Admin |
| `DELETE` | `/api/v1/routes/:id` | Eliminar ruta | 🔴 Admin / Super Admin |
| `PATCH` | `/api/v1/routes/:id/assign` | Asignar conductor a la ruta | 🔴 Admin / Super Admin |
| `PATCH` | `/api/v1/routes/:id/status` | Actualizar estado (In Progress/Completed) | 🔵 Driver Asignado |

### 📍 Puntos de Entrega (Waypoints)

| Método | Endpoint | Descripción | Nivel de Acceso |
| --- | --- | --- | --- |
| `PATCH` | `/api/v1/waypoints/:id/complete` | Completar entrega y **Subir Foto (POD)** | 🔵 Driver Asignado |
| `PUT` | `/api/v1/waypoints/:id` | Corregir dirección o datos del punto | 🔴 Admin / Super Admin |
