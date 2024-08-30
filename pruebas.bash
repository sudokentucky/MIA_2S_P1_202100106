# Crear el disco
mkdisk -size=2000 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/Disk3.mia

# Crear 4 particiones primarias de 500 MB cada una
fdisk -size=500 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/Disk3.mia -name="Primaria1"
fdisk -size=500 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/Disk3.mia -name="Primaria2"
fdisk -size=500 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/Disk3.mia -name="Primaria3"
fdisk -size=500 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/Disk3.mia -name="Primaria4"

# Intentar crear una partición extendida después de llenar todas las particiones primarias
fdisk -size=500 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/Disk3.mia -type=E -name="Extendida"

#/*Caso 2*/
# Crear el disco
mkdisk -size=1500 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/Disk4.mia

# Crear una partición extendida de 1000 MB
fdisk -size=1000 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/Disk4.mia -type=E -name="Extendida1"

# Crear particiones lógicas hasta llenar la partición extendida
fdisk -size=200 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/Disk3.mia -type=L -name="Logica1"
fdisk -size=200 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/Disk3.mia -type=L -name="Logica2"
fdisk -size=200 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/Disk3.mia -type=L -name="Logica3"
fdisk -size=200 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/Disk3.mia -type=L -name="Logica4"
fdisk -size=200 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/Disk3.mia -type=L -name="Logica5"
#caso 3*/
# Crear el disco
mkdisk -size=3000 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/Disk5.mia

# Crear una partición extendida de 1000 MB
fdisk -size=1000 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/Disk5.mia -type=E -name="Extendida1"

# Intentar crear otra partición extendida
fdisk -size=1000 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/Disk5.mia -type=E -name="Extendida2"
#/*caso 4*/
# Crear el disco
mkdisk -size=2000 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/Disk6.mia

# Crear una partición extendida de 1000 MB
fdisk -size=1000 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/Disk6.mia -type=E -name="Extendida1"

# Crear una partición lógica dentro de la extendida
fdisk -size=500 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/Disk6.mia -type=L -name="Logica1"

# Crear una partición primaria después de la extendida
fdisk -size=500 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/Disk6.mia -name="Primaria1"
#/*caso 5*/
# Crear el disco
mkdisk -size=2000 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/Disk7.mia

# Intentar crear una partición lógica sin una extendida
fdisk -size=500 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/Disk7.mia -type=L -name="Logica1"
#/*caso 6*/
# Crear el disco
mkdisk -size=3000 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/Disk8.mia

# Crear una partición extendida de 1000 MB
fdisk -size=1000 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/Disk8.mia -type=E -name="Extendida1"

# Crear particiones lógicas hasta llenar la extendida
fdisk -size=500 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/Disk8.mia -type=L -name="Logica1"
fdisk -size=500 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/Disk8.mia -type=L -name="Logica2"

# Crear particiones primarias hasta llenar el espacio restante
fdisk -size=500 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/Disk8.mia -name="Primaria1"
fdisk -size=500 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/Disk8.mia -name="Primaria2"
fdisk -size=500 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/Disk8.mia -name="Primaria3"
#/*caso 7*/
# Crear el disco
mkdisk -size=1500 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/Disk9.mia

# Crear una partición primaria de 1000 MB
fdisk -size=1000 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/Disk9.mia -name="Primaria1"

# Intentar crear una partición primaria de 600 MB (no debería permitirlo)
fdisk -size=600 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/Disk9.mia -name="Primaria2"
#/*caso 8*/
# Crear el disco
mkdisk -size=2500 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/Disk10.mia

# Crear una partición primaria de 500 MB
fdisk -size=500 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/Disk10.mia -name="Primaria1"

# Crear una partición extendida de 1500 MB
fdisk -size=1500 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/Disk10.mia -type=E -name="Extendida1"

# Crear particiones lógicas dentro de la extendida
fdisk -size=500 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/Disk10.mia -type=L -name="Logica1"
fdisk -size=500 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/Disk10.mia -type=L -name="Logica2"
fdisk -size=500 -unit=M -path=/home/keneth/Escritorio/Proyecto/Discos/Disk10.mia -type=L -name="Logica3"
