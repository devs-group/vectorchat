import { useState, useRef, useEffect } from 'react';
import { motion, AnimatePresence } from 'motion/react';
import { Send, Mic, BarChart3, X, Check } from 'lucide-react';

interface Message {
  id: string;
  text: string;
  isUser: boolean;
  timestamp: Date;
}

export function GlassChatWidget() {
  const [isOpen, setIsOpen] = useState(false);
  const [message, setMessage] = useState('');
  const [messages, setMessages] = useState<Message[]>([]);
  const [isTyping, setIsTyping] = useState(false);
  const [isRecording, setIsRecording] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const handleSend = async () => {
    if (!message.trim()) return;

    const userMessage: Message = {
      id: Date.now().toString(),
      text: message,
      isUser: true,
      timestamp: new Date(),
    };

    setMessages(prev => [...prev, userMessage]);
    setMessage('');
    setIsTyping(true);

    // Simulate AI response
    setTimeout(() => {
      const aiMessage: Message = {
        id: (Date.now() + 1).toString(),
        text: "I'm a demo AI assistant. I can help you with various tasks and questions.",
        isUser: false,
        timestamp: new Date(),
      };
      setMessages(prev => [...prev, aiMessage]);
      setIsTyping(false);
    }, 1200);
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const handleMicClick = () => {
    setIsRecording(true);
  };

  const handleAcceptRecording = () => {
    // Simulate converting voice to text
    const transcribedText = "This is a simulated voice message converted to text.";
    setMessage(transcribedText);
    setIsRecording(false);
  };

  const handleCancelRecording = () => {
    setIsRecording(false);
  };

  // Generate animated waveform bars
  const WaveformBars = () => {
    const bars = Array.from({ length: 40 }, (_, i) => (
      <motion.div
        key={i}
        className="bg-gray-300"
        style={{ width: '2px' }}
        animate={{
          height: [4, Math.random() * 20 + 4, 4],
        }}
        transition={{
          duration: 0.5 + Math.random() * 0.5,
          repeat: Infinity,
          repeatType: "reverse",
          delay: Math.random() * 0.5,
        }}
      />
    ));
    return <div className="flex items-center justify-center space-x-0.5 h-6">{bars}</div>;
  };

  return (
    <div className="fixed bottom-6 left-1/2 transform -translate-x-1/2 z-50">
      <AnimatePresence mode="wait">
        {!isOpen ? (
          // Collapsed floating input
          <motion.div
            key="collapsed"
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: 10 }}
            transition={{ duration: 0.2 }}
            className="relative"
          >
            <div 
              onClick={() => setIsOpen(true)}
              className="group cursor-pointer bg-black/60 backdrop-blur-xl border border-white/10 rounded-xl px-4 py-3 hover:bg-black/70 transition-all duration-200 min-w-[420px] max-w-[500px]"
            >
              <div className="flex items-center justify-between">
                <span className="text-gray-300 text-sm">Ask anything...</span>
                <div className="flex items-center space-x-2">
                  <button
                    onClick={handleMicClick}
                    className="hover:bg-white/10 p-1 rounded transition-colors"
                  >
                    <Mic className="w-4 h-4 text-gray-400 hover:text-gray-200" />
                  </button>
                  <BarChart3 className="w-4 h-4 text-gray-400" />
                </div>
              </div>
            </div>
          </motion.div>
        ) : (
          // Expanded chat interface
          <motion.div
            key="expanded"
            initial={{ opacity: 0, y: 20, height: 60 }}
            animate={{ opacity: 1, y: 0, height: 500 }}
            exit={{ opacity: 0, y: 10, height: 60 }}
            transition={{ duration: 0.3, ease: "easeOut" }}
            className="w-[420px] bg-black/60 backdrop-blur-xl border border-white/10 rounded-xl overflow-hidden relative"
          >
            {/* Close button */}
            <button
              onClick={() => setIsOpen(false)}
              className="absolute top-3 right-3 z-10 p-1.5 hover:bg-white/10 rounded-lg transition-colors"
            >
              <X className="w-4 h-4 text-gray-400 hover:text-gray-200" />
            </button>

            {/* Messages */}
            <div className="flex-1 overflow-y-auto p-4 space-y-3 h-[420px] pt-12">
              {messages.length === 0 && (
                <div className="text-center text-gray-400 mt-8">
                  <p className="text-sm">Start a conversation</p>
                </div>
              )}
              
              {messages.map((msg) => (
                <motion.div
                  key={msg.id}
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.2 }}
                  className={`flex ${msg.isUser ? 'justify-end' : 'justify-start'}`}
                >
                  <div
                    className={`max-w-[320px] px-3 py-2 rounded-lg text-sm ${
                      msg.isUser
                        ? 'bg-white/20 text-white'
                        : 'bg-white/10 text-gray-200'
                    }`}
                  >
                    {msg.text}
                  </div>
                </motion.div>
              ))}

              {isTyping && (
                <motion.div
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  className="flex justify-start"
                >
                  <div className="bg-white/10 px-3 py-2 rounded-lg">
                    <div className="flex space-x-1">
                      <motion.div
                        animate={{ opacity: [0.4, 1, 0.4] }}
                        transition={{ duration: 1.4, repeat: Infinity, delay: 0 }}
                        className="w-1.5 h-1.5 bg-gray-400 rounded-full"
                      />
                      <motion.div
                        animate={{ opacity: [0.4, 1, 0.4] }}
                        transition={{ duration: 1.4, repeat: Infinity, delay: 0.2 }}
                        className="w-1.5 h-1.5 bg-gray-400 rounded-full"
                      />
                      <motion.div
                        animate={{ opacity: [0.4, 1, 0.4] }}
                        transition={{ duration: 1.4, repeat: Infinity, delay: 0.4 }}
                        className="w-1.5 h-1.5 bg-gray-400 rounded-full"
                      />
                    </div>
                  </div>
                </motion.div>
              )}
              <div ref={messagesEndRef} />
            </div>

            {/* Input */}
            <div className="p-4 border-t border-white/10">
              {isRecording ? (
                // Recording interface with waveform
                <div className="flex items-center space-x-3">
                  <button
                    onClick={handleCancelRecording}
                    className="p-1.5 hover:bg-white/10 rounded-lg transition-colors"
                  >
                    <X className="w-4 h-4 text-gray-400 hover:text-gray-200" />
                  </button>
                  <div className="flex-1">
                    <WaveformBars />
                  </div>
                  <button
                    onClick={handleAcceptRecording}
                    className="p-1.5 hover:bg-white/10 rounded-lg transition-colors"
                  >
                    <Check className="w-4 h-4 text-gray-400 hover:text-gray-200" />
                  </button>
                </div>
              ) : (
                // Normal input interface
                <div className="flex items-center space-x-3">
                  <div className="flex-1 relative">
                    <input
                      type="text"
                      value={message}
                      onChange={(e) => setMessage(e.target.value)}
                      onKeyPress={handleKeyPress}
                      placeholder="Type your message..."
                      className="w-full bg-transparent text-gray-200 placeholder-gray-500 text-sm focus:outline-none"
                      autoFocus
                    />
                  </div>
                  <div className="flex items-center space-x-2">
                    <button
                      onClick={handleMicClick}
                      className="hover:bg-white/10 p-1 rounded transition-colors"
                    >
                      <Mic className="w-4 h-4 text-gray-400 hover:text-gray-200" />
                    </button>
                    {message.trim() ? (
                      <button
                        onClick={handleSend}
                        className="w-4 h-4 text-gray-300 hover:text-white transition-colors"
                      >
                        <Send className="w-4 h-4" />
                      </button>
                    ) : (
                      <BarChart3 className="w-4 h-4 text-gray-400" />
                    )}
                  </div>
                </div>
              )}
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}