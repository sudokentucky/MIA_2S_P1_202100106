import React, { useState, useCallback, useEffect, useRef } from "react";
import FileInput from "./FileInput";
import Message from "./Message";

export default function CommandExecution() {
  const [inputText, setInputText] = useState("");
  const [outputText, setOutputText] = useState("");
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState("");
  const [messageType, setMessageType] = useState<"success" | "error" | "info" | "">("");
  const [lineCount, setLineCount] = useState(1);

  const textareaRef = useRef<HTMLTextAreaElement>(null); 
  const lineCounterRef = useRef<HTMLDivElement>(null); 

  const showMessage = (text: string, type: "success" | "error" | "info") => {
    setMessage(text);
    setMessageType(type);
    setTimeout(() => {
      setMessage("");
      setMessageType("");
    }, 5000);
  };

  useEffect(() => {
    const lines = inputText.split("\n").length;
    setLineCount(lines);
  }, [inputText]);

  const syncScroll = () => {
    if (textareaRef.current && lineCounterRef.current) {
      lineCounterRef.current.scrollTop = textareaRef.current.scrollTop;
    }
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
    <div className="flex flex-col min-h-screen" style={{ backgroundColor: "#E6E6FA" }}>
      <div className="flex-grow flex items-center justify-center p-4">
        {/* Reducir el tamaño del panel */}
        <div className="w-full max-w-5xl p-8 bg-beige rounded-lg shadow-lg transition-all hover:shadow-xl transform hover:scale-105 duration-300 ease-in-out" style={{ backgroundColor: "#F5F5DC" }}>
          <h1 className="text-3xl font-bold mb-6 text-center text-gray-800">
            Sistema de archivos ext2
          </h1>

          <div className="mb-4 relative">
            <label className="block text-base font-medium text-gray-700 mb-2" htmlFor="input-text">
              Entrada de comando o archivo de texto
            </label>

            <div className="flex">
              {/* Line counter */}
              <div
                ref={lineCounterRef}
                className="line-numbers bg-gray-200 p-2 rounded-l-md text-sm text-right overflow-hidden"
                style={{ height: 'auto', minHeight: '200px', maxHeight: '300px' }} 
              >
                {Array.from({ length: lineCount }, (_, i) => i + 1).map((line) => (
                  <div key={line}>{line}</div>
                ))}
              </div>

              {/* Input text area */}
              <textarea
                id="input-text"
                ref={textareaRef}
                className="w-full min-h-56 p-3 border border-gray-300 rounded-r-md resize-none shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition duration-200 text-sm overflow-y-auto"
                value={inputText}
                onChange={(e) => setInputText(e.target.value)}
                onScroll={syncScroll}
                placeholder="Ingrese comandos aquí..."
                disabled={loading}
                style={{ height: "auto", minHeight: "200px", fontSize: '13px', whiteSpace: 'pre' }}
              />
            </div>
          </div>

          <div className="mb-4">
            <label className="block text-base font-medium text-gray-700 mb-2" htmlFor="output-text">
              Resultado de la ejecución
            </label>
            <textarea
              id="output-text"
              className="w-full min-h-56 p-3 border border-gray-300 rounded-md resize-none bg-gray-100 shadow-sm focus:outline-none focus:ring-2 focus:ring-green-500 text-sm"
              value={outputText}
              readOnly
              placeholder="Resultado de la ejecución aparecerá aquí..."
              style={{ fontFamily: '"Courier New", monospace', fontSize: '13px', minHeight: '200px' }}
            />
          </div>

          <div className="flex justify-between items-center space-x-4">
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

          {loading && (
            <div className="mt-4 flex justify-center items-center">
              <div className="loader ease-linear rounded-full border-4 border-t-4 border-gray-200 h-6 w-6 mb-4"></div>
              <span className="text-blue-500 ml-2">Procesando...</span>
            </div>
          )}
        </div>
      </div>

    </div>
  );
}
