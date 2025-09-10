import { GlassChatWidget } from "./components/GlassChatWidget";

export default function App() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-blue-900 to-slate-800 flex items-center justify-center relative">
      {/* Subtle background pattern */}
      <div className="absolute inset-0 opacity-20">
        <div className="absolute inset-0" style={{
          backgroundImage: `radial-gradient(circle at 25% 25%, rgba(120, 119, 198, 0.3) 0%, transparent 50%),
                           radial-gradient(circle at 75% 75%, rgba(99, 102, 241, 0.2) 0%, transparent 50%)`
        }}></div>
      </div>

      {/* Main content */}
      <div className="relative z-10 text-center px-6 max-w-2xl">
        <h1 className="text-3xl text-white mb-4">
          Minimal AI Chat Interface
        </h1>
        <p className="text-gray-300 mb-8 leading-relaxed">
          A clean, minimalistic chat widget with glass material design. 
          Click the input below to start chatting with AI.
        </p>
        <div className="flex justify-center gap-6 text-sm text-gray-400">
          <span>• Dark Theme</span>
          <span>• Glass Material</span>
          <span>• Minimal Design</span>
          <span>• Smooth Animations</span>
        </div>
      </div>

      {/* Glass Chat Widget */}
      <GlassChatWidget />
    </div>
  );
}