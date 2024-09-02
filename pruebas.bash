# Crear un disco de 500 MB
mkdisk -size=500 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/ExampleDisk.mia
# Crear 3 particiones primarias de 50 MB cada una
fdisk -size=50 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/ExampleDisk.mia -name="Primaria1"
fdisk -size=50 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/ExampleDisk.mia -name="Primaria2"
fdisk -size=50 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/ExampleDisk.mia -name="Primaria3"

# Crear una partición extendida de 200 MB
fdisk -size=200 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/ExampleDisk.mia -type=E -name="Extendida"
# Crear 3 particiones lógicas de 50 MB cada una dentro de la extendida
fdisk -size=50 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/ExampleDisk.mia -type=L -name="Logica1"
fdisk -size=50 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/ExampleDisk.mia -type=L -name="Logica2"
fdisk -size=50 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/ExampleDisk.mia -type=L -name="Logica3"
# Montar particiones primarias
mount -path=/home/keneth/Escritorio/Proyecto/Discos/ExampleDisk.mia -name="Primaria1"
mount -path=/home/keneth/Escritorio/Proyecto/Discos/ExampleDisk.mia -name="Primaria2"
mount -path=/home/keneth/Escritorio/Proyecto/Discos/ExampleDisk.mia -name="Primaria3"

rep -id=061A -path=/home/keneth/Escritorio/Proyecto/Reports/ExampleDisk_MBR.png -name=mbr
rep -id=061A -path=/home/keneth/Escritorio/Proyecto/Reports/ExampleDisk_Disk.png -name=disk
