import Image from "next/image";

export default function Home() {
  return (
    <div className="bg-background text-on-background min-h-screen flex flex-col font-inter">
      <nav className="fixed top-0 w-full z-50 bg-white/20 dark:bg-gray-900/20 backdrop-blur-[20px] text-blue-700 dark:text-blue-400 tracking-[-0.02em] text-sm font-medium border-b border-white/10">
        <div className="flex items-center justify-between px-6 py-4 max-w-7xl mx-auto">
          <a className="text-xl font-bold tracking-tighter text-gray-900 dark:text-white scale-95 active:scale-100 transition-transform" href="#">Digital Atelier</a>
          <div className="hidden md:flex items-center space-x-8">
            <a className="text-blue-700 dark:text-blue-400 border-b-2 border-blue-700 pb-1 hover:text-gray-900 dark:hover:text-white transition-colors duration-300 scale-95 active:scale-100" href="#">Mac</a>
            <a className="text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white transition-colors duration-300 scale-95 active:scale-100" href="#">iPhone</a>
            <a className="text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white transition-colors duration-300 scale-95 active:scale-100" href="#">iPad</a>
            <a className="text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white transition-colors duration-300 scale-95 active:scale-100" href="#">Watch</a>
            <a className="text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white transition-colors duration-300 scale-95 active:scale-100" href="#">Vision</a>
            <a className="text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white transition-colors duration-300 scale-95 active:scale-100" href="#">Support</a>
          </div>
          <div className="flex items-center space-x-4">
            <button className="text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white transition-colors duration-300 scale-95 active:scale-100">
              <span className="material-symbols-outlined" data-icon="shopping_bag">shopping_bag</span>
            </button>
          </div>
        </div>
      </nav>
      <main className="flex-grow flex flex-col gap-8 md:gap-12">
        <section className="relative w-full h-[921px] md:h-screen flex items-center justify-center overflow-hidden">
          <img alt="Sleek silver MacBook Pro" className="absolute inset-0 w-full h-full object-cover brightness-[0.7] contrast-[1.1]" src="https://lh3.googleusercontent.com/aida-public/AB6AXuC8shpRMiA4LDJyY2iE9iwGMoKUtP-OmCRkkev1DnL_deFmdb9KWNemaZpEx3_BVGBNTLyWSw-p346zF-R0rn5j_hBXYe2N7Mr8-IB8aoNAVjBc6viyGmQ6N2UWQ0ijoQU2mNfki1kGSWgkimsTOQC3TmPTfME5c_Wr4OF_R2gI_QIRevVO5NabEGs_XFTflm9T7eQ4rNbHxtcwzEn2FZmWejqgZa3zGEF2Q54LqXxB_EuJ7APFWuOblMo49cuqOdJXEJPuhgI-T5w"/>
          <div className="z-10 flex flex-col items-center gap-6 max-w-4xl px-6 text-center text-white">
            <h2 className="text-5xl md:text-8xl font-bold tracking-[-0.03em] drop-shadow-2xl">MacBook Pro</h2>
            <p className="text-xl md:text-2xl max-w-2xl font-medium tracking-tight opacity-90">Mind-blowing. Head-turning.</p>
            <div className="flex flex-row items-center gap-6 mt-4">
              <button className="bg-primary text-on-primary px-10 py-3.5 rounded-full font-semibold tracking-wide hover:bg-primary-container transition-colors duration-300 shadow-xl">Buy</button>
              <a className="text-white font-semibold tracking-wide hover:opacity-80 transition-opacity duration-300 border-b-[1.5px] border-white pb-1 flex items-center gap-1" href="#">Learn more <span className="material-symbols-outlined align-middle text-sm">chevron_right</span></a>
            </div>
          </div>
        </section>
        <section className="max-w-7xl mx-auto px-6 w-full grid grid-cols-1 md:grid-cols-3 gap-6 mb-12">
          <div className="aspect-square bg-surface-container-lowest rounded-xl overflow-hidden relative group shadow-[0_20px_40px_rgba(26,28,29,0.04)] flex flex-col">
            <div className="p-8 pb-4 text-center z-10">
              <h3 className="text-2xl font-bold tracking-tight text-on-surface">iPhone 15 Pro</h3>
              <p className="text-sm text-secondary mt-1 font-medium">Titanium. So strong. So light. So Pro.</p>
              <div className="flex flex-row items-center justify-center gap-4 mt-3">
                <button className="bg-primary text-on-primary px-5 py-1.5 text-sm rounded-full font-medium hover:bg-primary-container transition-colors">Buy</button>
                <a className="text-primary font-medium text-sm hover:text-primary-container transition-colors border-b border-primary pb-0.5" href="#">Learn more</a>
              </div>
            </div>
            <div className="flex-grow relative mt-4">
              <img alt="iPhone 15 Pro" className="absolute inset-0 w-full h-full object-cover object-center group-hover:scale-105 transition-transform duration-700 ease-out" src="https://lh3.googleusercontent.com/aida-public/AB6AXuBRGrLyo00BwGx13cOlWiwcJF8SL3MfZIm39QDOAphoDQJhn-baUlCAtXQEYU5ZpLmB-zCghc1Wp7oKMoNgWKwCuy2mYAIFLbF15LtXg2MgkFWxcEvjN0SKC_Zw8rDcm9-vMW7oyAwXGlI_IJT6sIcY_zLbUCH9Y8DhKuNtyxvQrUcs2qavPE7gKrRgjNvDmOPvODZxG9lwAtccgtM2pD7ZDv3yolTslxatj1K3Ur7wKZxsNa1TFwJn0sCFjFcVZnAN5slNGEfE4LY"/>
            </div>
          </div>
          <div className="aspect-square bg-surface-container-low rounded-xl overflow-hidden relative group shadow-[0_20px_40px_rgba(26,28,29,0.04)] flex flex-col">
            <div className="p-8 pb-4 text-center z-10">
              <h3 className="text-2xl font-bold tracking-tight text-on-surface">iPad Pro</h3>
              <p className="text-sm text-secondary mt-1 font-medium">Supercharged by M2.</p>
              <div className="flex flex-row items-center justify-center gap-4 mt-3">
                <button className="bg-primary text-on-primary px-5 py-1.5 text-sm rounded-full font-medium hover:bg-primary-container transition-colors">Buy</button>
                <a className="text-primary font-medium text-sm hover:text-primary-container transition-colors border-b border-primary pb-0.5" href="#">Learn more</a>
              </div>
            </div>
            <div className="flex-grow relative mt-4 px-6 pb-6">
              <img alt="iPad Pro" className="absolute inset-0 w-full h-full object-contain object-bottom group-hover:scale-105 transition-transform duration-700 ease-out" src="https://lh3.googleusercontent.com/aida-public/AB6AXuB4oAYJAancB6POIM7Obv2qVM9Si0s15MBeIz9P5wIZqznugg4jE3tjQ2wIec-2QDEu5mbrdaXnrVrEagHigUVHOSQb27ckyb6v4_SpMgw352Esbf0g8PVB-ntvmc1GbBG4eLXhuS7h8xm0EacDBhYr4lKzkFbN94RVLAdkYaZEfe55kRdq0ty-3D1dyXExAjSHbwN4VI93WT8EG_OyTzy1hCU_IUL9VAWX9-TFI-XYzh0GQPFJRrzk4o7x9fjyhKKAEO-I4IuI2H0"/>
            </div>
          </div>
          <div className="aspect-square bg-surface-container-highest rounded-xl overflow-hidden relative group shadow-[0_20px_40px_rgba(26,28,29,0.04)] flex flex-col">
            <div className="p-8 pb-4 text-center z-10">
              <h3 className="text-2xl font-bold tracking-tight text-on-surface">Watch</h3>
              <p className="text-sm text-secondary mt-1 font-medium uppercase tracking-widest">Series 9</p>
              <div className="flex flex-row items-center justify-center gap-4 mt-3">
                <button className="bg-primary text-on-primary px-5 py-1.5 text-sm rounded-full font-medium hover:bg-primary-container transition-colors">Buy</button>
                <a className="text-primary font-medium text-sm hover:text-primary-container transition-colors border-b border-primary pb-0.5" href="#">Learn more</a>
              </div>
            </div>
            <div className="flex-grow relative mt-4 px-10 pb-10">
              <img alt="Apple Watch" className="absolute inset-0 w-full h-full object-contain object-bottom group-hover:scale-105 transition-transform duration-700 ease-out" src="https://lh3.googleusercontent.com/aida-public/AB6AXuDHjNyrHAt2UF51j_WmY-kAtgbRTkBbJhEAWyD-t0I9yFJTi1AnwatgrAEpB3GiudtaN_2J1t04TKzz0N988SISxZ6K7M4_q5vhHSUZxLNEhPhe9IoZrKQIQ3epB5iC9Ed9luth3TCjqNOjRhYmzgWOW95NH_58uJfPEfSH-giFhhL1YGBRFpS0j97vAJo5f-CJC31GppM_nUu6X26AUJG_ytyGGhK63my-P0f2p01aqCeiiIviXlZxXihzFu_xSx5NvY2s9R-9Q4A"/>
            </div>
          </div>
        </section>
      </main>
      <footer className="w-full py-12 bg-gray-50 dark:bg-gray-950 text-blue-700 dark:text-blue-400 text-xs leading-relaxed border-t border-surface-variant">
        <div className="max-w-7xl mx-auto px-6 grid grid-cols-1 md:grid-cols-2 gap-8">
          <div className="flex flex-col space-y-4">
            <span className="text-gray-900 dark:text-white font-semibold">© 2024 Digital Atelier. Designed for Apple Enthusiasts.</span>
          </div>
          <div className="flex flex-wrap gap-6 md:justify-end">
            <a className="text-gray-500 dark:text-gray-400 hover:underline underline-offset-8 transition-all opacity-80 hover:opacity-100" href="#">Privacy Policy</a>
            <a className="text-gray-500 dark:text-gray-400 hover:underline underline-offset-8 transition-all opacity-80 hover:opacity-100" href="#">Terms of Use</a>
            <a className="text-gray-500 dark:text-gray-400 hover:underline underline-offset-8 transition-all opacity-80 hover:opacity-100" href="#">Sales Policy</a>
            <a className="text-gray-500 dark:text-gray-400 hover:underline underline-offset-8 transition-all opacity-80 hover:opacity-100" href="#">Legal</a>
            <a className="text-gray-500 dark:text-gray-400 hover:underline underline-offset-8 transition-all opacity-80 hover:opacity-100" href="#">Site Map</a>
          </div>
        </div>
      </footer>
    </div>
  );
}
