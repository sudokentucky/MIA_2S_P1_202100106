import React from "react";

// Message component to display messages
interface MessageProps {
  text: string;
  type: "success" | "error" | "info" | "";
}

const Message: React.FC<MessageProps> = ({ text, type }) => {
  if (!text) return null;

  const messageClass = type === "success"
    ? "bg-green-100 text-green-700"
    : type === "error"
    ? "bg-red-100 text-red-700"
    : "bg-blue-100 text-blue-700";

  return (
    <div className={`mt-4 p-2 text-center rounded-md ${messageClass} transition-opacity duration-500 ease-in-out`}>
      {text}
    </div>
  );
};

export default Message;
