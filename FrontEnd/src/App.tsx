import React, { useState, useCallback } from "react";
import FileInput from "./FileInput";
import Message from "./Message";

export default function CommandExecution() {
  const [inputText, setInputText] = useState("");
  const [outputText, setOutputText] = useState("");
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState("");
  const [messageType, setMessageType] = useState<"success" | "error" | "info" | "">("");

  const showMessage = (text: string, type: "success" | "error" | "info") => {
    setMessage(text);
    setMessageType(type);
    setTimeout(() => {
      setMessage("");
      setMessageType("");
    }, 5000);
  };

  const handleExecute = useCallback(async () => {
    if (!inputText.trim()) {
      showMessage("El área de texto está vacía. Por favor, ingrese un comando o cargue un archivo.", "error");
      return;
    }

    setLoading(true);
    try {
      const response = await fetch("http://localhost:3000/analyze", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ command: inputText }),
      });

      if (!response.ok) {
        throw new Error("Error en la red o en la respuesta del servidor");
      }

      const data = await response.json();
      const results = data.results.join("\n");
      setOutputText(results);
      showMessage("Ejecución completada con éxito", "success");

    } catch (error) {
      console.error("Error:", error);
      if (error instanceof Error) {
        setOutputText(`Error: ${error.message}`);
        showMessage(`Error en la ejecución: ${error.message}`, "error");
      } else {
        setOutputText("Error desconocido");
        showMessage("Error en la ejecución: Error desconocido", "error");
      }
    } finally {
      setLoading(false);
    }
  }, [inputText]);

  const handleReset = () => {
    setInputText("");
    setOutputText("");
    showMessage("Campos limpiados correctamente", "info");
  };

  return (
    <div className="flex flex-col min-h-screen bg-gray-100 font-inter">
      <div className="flex-grow flex items-center justify-center p-4">
        <div className="w-full max-w-3xl p-8 bg-white rounded-lg shadow-md transition-all hover:shadow-xl transform hover:scale-105 duration-300 ease-in-out">
          <h1 className="text-3xl font-bold mb-6 text-center text-gray-800">
            Manejo e Implementacion de Archivos
          </h1>

          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 mb-2" htmlFor="input-text">
              Entrada de comando o archivo de texto
            </label>
            <textarea
              id="input-text"
              className="w-full h-48 p-3 border border-gray-300 rounded-md resize-none shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition duration-200"
              value={inputText}
              onChange={(e) => setInputText(e.target.value)}
              placeholder="Ingrese comandos aquí..."
              disabled={loading}
            />
          </div>

          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 mb-2" htmlFor="output-text">
              Resultado de la ejecución
            </label>
            <textarea
              id="output-text"
              className="w-full h-48 p-3 border border-gray-300 rounded-md resize-none bg-gray-100 shadow-sm focus:outline-none focus:ring-2 focus:ring-green-500"
              value={outputText}
              readOnly
              placeholder="Resultado de la ejecución aparecerá aquí..."
            />
          </div>

          <div className="flex justify-between items-center">
            <FileInput onFileChange={setInputText} showMessage={showMessage} loading={loading} />

            <button
              onClick={handleExecute}
              className={`px-4 py-2 rounded-md text-white focus:outline-none flex items-center transition-all duration-200 ${
                loading
                  ? "bg-gray-400"
                  : "bg-green-500 hover:bg-green-600 focus:ring-2 focus:ring-green-500"
              }`}
              disabled={loading}
            >
              <span className="mr-2">▶️</span>
              {loading ? "Ejecutando..." : "Ejecutar"}
            </button>

            <button
              onClick={handleReset}
              className="px-4 py-2 bg-red-500 text-white rounded-md hover:bg-red-600 focus:outline-none focus:ring-2 focus:ring-red-500 flex items-center transition-all duration-200"
              disabled={loading}
            >
              🧹 Limpiar
            </button>
          </div>

          <Message text={message} type={messageType} />

          {loading && <div className="mt-4 text-center text-blue-500">Procesando...</div>}
        </div>
      </div>

      <footer className="py-4 text-center text-sm text-gray-500">
         Keneth Lopez - 202100106
      </footer>
    </div>
  );
}
