# MIA_2S_P1_202100106
## Descripcion del Proyecto

Este proyecto ofrece una interfaz de usuario creada con React para gestionar y ejecutar comandos de un sistema de archivos ext2 y ext3 de manera simulada. El usuario puede escribir sus comandos directamente en el área de texto o cargar un archivo, y la aplicación se encarga de enviar dicha información a un servidor para su análisis. El resultado de la ejecución se muestra en tiempo real, acompañado de un contador de líneas, indicadores de carga y notificaciones de éxito, error o información. Además, se incluyen opciones para reiniciar el campo de entrada y cargar nuevos archivos de forma fácil, brindando una experiencia de usuario fluida e intuitiva.

## Características principales

- **Ejecución de comandos**: Ingresa comandos manualmente o a través de un archivo para procesarlos en el servidor.
- **Retroalimentación en tiempo real**: Muestra los resultados en pantalla y notifica al usuario con mensajes de éxito, error o información.
- **Contador de líneas**: Facilita la edición de los comandos al mostrar la numeración de cada línea.
- **Limpieza rápida**: Botón dedicado para resetear el texto de entrada y los resultados de salida.
- **Carga de archivos**: Posibilidad de subir un archivo de texto para ejecutar varios comandos a la vez.
- **Interfaces intuitivas**: Uso de React y TailwindCSS para lograr un diseño responsivo y agradable.

## Tecnologías y dependencias

Este proyecto se divide en dos partes: **Frontend** (React + Vite + TailwindCSS) y **Backend** (Go + Fiber).

### Frontend

- **React** / **React DOM**: Biblioteca para construir la interfaz de usuario.  
- **Vite**: Herramienta de desarrollo y empaquetado rápido.  
- **TailwindCSS**: Framework CSS para un diseño responsivo con clases utilitarias.  
- **TypeScript**: Aporta tipado estático y mayor robustez en el desarrollo.  
- **ESLint**: Análisis estático para mantener un código consistente.  
- **PostCSS** / **Autoprefixer**: Procesamiento de CSS para compatibilidad entre navegadores.  

#### Instalacion React & Vite

```bash
#Utiliar el empaquetador de preferencia en este caso npm
npm install -D vite
#Crear el proyecto
npm create vite@latest
#Entrar a la carpeta creada e instalar las dependencias
npm install
#Iniciar el servidor
npm run dev
```

#### Tailwindcss

```bash
npm install -D tailwindcss postcss autoprefixer
npx tailwindcss init -p
```

Se va a crear un archivo tailwind.config.js en la raíz del proyecto. En este archivo se pueden configurar los estilos de tailwindcss. Se debe agregar el siguiente código:

```bash
#Archivo tailwind.config.js
content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
```

Por último se debe importar tailwindcss en el archivo index.css:

```bash
/* Archivo index.css */
@tailwind base;
@tailwind components;
@tailwind utilities;
```

### Backend

- **Go (Golang)**: Lenguaje de programación de alto rendimiento para la lógica del lado del servidor.  
- **Fiber**: Framework web rápido y minimalista para construir APIs.  

#### Instalacion Fiber

```bash
#Iniciar el modulo de Go en el proyecto
go mod init backend
#Se instala el framework de Go (Fiber)
go get -u github.com/gofiber/fiber/v2
#Iniciar el servidor de desarrollo con el siguiente comando
go run main.go
```

### Ejecucion del Proyecto

- Luego de iniciar los entornos tanto de frontend como backend,abra <http://localhost:3000> para el servidor (Esto util para probar endpoints o ver el backend sin la interfaz grafica) y <http://localhost:5173> para ver la aplicación en el navegador.

---
