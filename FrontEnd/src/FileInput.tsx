import React, { useRef } from "react";

interface FileInputProps {
  onFileChange: (fileContent: string) => void;
  showMessage: (text: string, type: "success" | "error" | "info") => void;
  loading: boolean;
}
// FileInput component to handle file input
const FileInput: React.FC<FileInputProps> = ({ onFileChange, showMessage, loading }) => {
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file && file.type === "text/plain") {
      const reader = new FileReader();
      reader.onload = (e) => {
        const content = e.target?.result as string;
        onFileChange(content);
        showMessage("Archivo cargado exitosamente", "success");
      };
      reader.onerror = () => {
        showMessage("Error al leer el archivo", "error");
      };
      reader.readAsText(file);
    } else {
      showMessage("Por favor, seleccione un archivo de texto (.txt)", "error");
    }
  };

  return (
    <div>
      <input
        type="file"
        ref={fileInputRef}
        onChange={handleFileChange}
        className="hidden"
        accept=".txt"
      />
      <button
        onClick={() => fileInputRef.current?.click()}
        className="px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 flex items-center"
        disabled={loading}
      >
        <span className="mr-2">📁</span>
        Examinar
      </button>
    </div>
  );
};

export default FileInput;
