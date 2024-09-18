# Crear un disco de 500 MB
mkdisk -size=500 -unit=M -path=/home/sudokentucky/Escritorio/Archivos/Pruebas/Disks/ExampleDisk.mia 
# Crear 3 particiones primarias de 50 MB cada una
fdisk -size=50 -unit=M -path=/home/sudokentucky/Escritorio/Archivos/Pruebas/Disks/ExampleDisk.mia  -name="Primaria1"
fdisk -size=50 -unit=M -path=/home/sudokentucky/Escritorio/Archivos/Pruebas/Disks/ExampleDisk.mia  -name="Primaria2"
fdisk -size=50 -unit=M -path=/home/sudokentucky/Escritorio/Archivos/Pruebas/Disks/ExampleDisk.mia  -name="Primaria3"

# Crear una partición extendida de 200 MB
fdisk -size=200 -unit=M -path=/home/sudokentucky/Escritorio/Archivos/Pruebas/Disks/ExampleDisk.mia  -type=E -name="Extendida"
# Crear 3 particiones lógicas de 50 MB cada una dentro de la extendida
fdisk -size=50 -unit=M -path=/home/sudokentucky/Escritorio/Archivos/Pruebas/Disks/ExampleDisk.mia  -type=L -name="Logica1"
fdisk -size=50 -unit=M -path=/home/sudokentucky/Escritorio/Archivos/Pruebas/Disks/ExampleDisk.mia -type=L -name="Logica2"
fdisk -size=50 -unit=M -path=/home/sudokentucky/Escritorio/Archivos/Pruebas/Disks/ExampleDisk.mia  -type=L -name="Logica3"
# Montar particiones primarias
mount -path=/home/sudokentucky/Escritorio/Archivos/Pruebas/Disks/ExampleDisk.mia  -name="Primaria1"
mount -path=/home/sudokentucky/Escritorio/Archivos/Pruebas/Disks/ExampleDisk.mia  -name="Primaria2"
mount -path=/home/sudokentucky/Escritorio/Archivos/Pruebas/Disks/ExampleDisk.mia  -name="Primaria3"
mkfs -id=061A -type=full 
login -user=root -pass=123 -id=061A
#Crear grupos
mkgrp -name=usuarios
mkgrp -name=administradores
mkgrp -name=admin
mkgrp -name=home
mkgrp -name=prueba
mkgrp -name=userz

#Crear usuarios
mkusr -user=keni -pass=123 -grp=usuarios
mkusr -user=keneth -pass=123 -grp=admin
#usuarios para usuarios
mkusr -user=prueba -pass=123 -grp=usuarios
mkusr -user=prueba2 -pass=123 -grp=usuarios
mkusr -user=k99 -pass=123 -grp=prueba
mkusr -user=keni -pass=123 -grp=home
#Eliminar un grupo
rmgrp -name=usuarios
#Eliminar un usuario
rmusr -usr=keni
rmusr -usr=keneth
#cambiar el usuario de grupo
chgrp -usr=keni -grp=admin
mkfile -size=15 -path=/home/user/docs/a.txt -r

mkgrp -name=usuarios // Debe dar error porque ya existe
mkgrp -name=administradores

mkusr -user=keni -pass=123 -grp=usuarios // Debe dar error porque ya existe
mkusr -user=keneth -pass=123 -grp=administradores
mkusr -usr=prueba -pass=123 -grp=usuarios //probar la asignacion de punteros a usuarios con mas usuarios

logout
login -user=keni -pass=123 -id=061A // Usuario aun no creado debe dar error
# Crea el archivo 'a.txt' en la ruta especificada
# Las carpetas '/home/user/docs' se crean automáticamente si no existen
# El tamaño del archivo será de 15 bytes, con el contenido "012345678901234"

# Crea el archivo 'archivo 1.txt' en la carpeta 'mis documentos'
# No se crean carpetas adicionales, y el archivo tendrá 0 bytes de tamaño
mkdir -path="/home"
mkdir -path="/home/usac"
mkdir -path="/home/work"
mkdir -path="/home/usac/mia"
mkfile -path="/home/usac/mia/1.txt"
mkfile -path="/home/mis documentos/archivo 1.txt"

mkfile -size=10 -path=/home/user/docs/a.txt
# Error: Las carpetas padres no existen y no se utilizó el parámetro -r para crearlas.

rep -id=061A -path=/home/sudokentucky/Escritorio/Archivos/Pruebas/Reps/ExampleDisk_MBR.png -name=mbr
rep -id=061A -path=/home/sudokentucky/Escritorio/Archivos/Pruebas/Reps/ExampleDisk_Disk.png -name=disk
rep -id=061A -path=/home/sudokentucky/Escritorio/Archivos/Pruebas/Reps/ExampleDisk_Sb.png -name=sb
rep -id=061A -path=/home/sudokentucky/Escritorio/Archivos/Pruebas/Reps/ExampleDisk_block.png -name=block
rep -id=061A -path=/home/sudokentucky/Escritorio/Archivos/Pruebas/Reps/ExampleDisk_bm_block.txt -name=bm_block
rep -id=061A -path=/home/sudokentucky/Escritorio/Archivos/Pruebas/Reps/Example_bm_inode.txt -name=bm_inode
rep -id=061A -path=/home/sudokentucky/Escritorio/Archivos/Pruebas/Reps/Example_bm_block.txt -name=bm_block
rep -id=061A -path=/home/sudokentucky/Escritorio/Archivos/Pruebas/Reps/Example_inode.png -name=inode
rmdisk -path=/home/sudokentucky/Escritorio/Archivos/Pruebas/Disks/ExampleDisk.mia 