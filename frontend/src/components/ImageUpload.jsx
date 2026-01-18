import { useRef } from 'react';

function ImageUpload({ imageUrl, uploading, onFileSelect, onRemove }) {
  // As we need a customized <input />, we need to use ref to get the <input />.(since it's hidden)
  const fileInputRef = useRef(null);

  // upload image handler
  const handleFileChange = (e) => {
    const selectedFile = e.target.files[0];
    // only support png and jpg
    const allowedTypes = ['image/png', 'image/jpeg'];
    if (selectedFile && allowedTypes.includes(selectedFile.type)) {
      onFileSelect(selectedFile);
    }
    fileInputRef.current.value = '';
  };

  const handleClick = () => {
    fileInputRef.current.click();
  };

  return (
    <div className="w-full">
      <input
        type="file"
        ref={fileInputRef}
        onChange={handleFileChange}
        accept="image/png,image/jpeg"
        className="hidden"
      />

      {!imageUrl && !uploading && (
        <button
          onClick={handleClick}
          className="w-full aspect-square border-2 border-dashed border-gray-300 rounded-lg flex flex-col items-center justify-center gap-3 hover:border-purple-400 hover:bg-purple-50 transition-colors"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            className="h-12 w-12 text-gray-400"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={1.5}
              d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
            />
          </svg>
          <span className="text-gray-500">Click to select an image</span>
        </button>
      )}

      {uploading && (
        <div className="w-full aspect-square border-2 border-dashed border-gray-300 rounded-lg flex flex-col items-center justify-center gap-3">
          <div className="w-10 h-10 border-4 border-purple-500 border-t-transparent rounded-full animate-spin"></div>
          <p className="text-sm text-gray-500">Uploading...</p>
        </div>
      )}

      {imageUrl && !uploading && (
        <div className="relative rounded-lg overflow-hidden">
          <img
            src={imageUrl}
            alt="Preview"
            className="w-full aspect-square object-cover"
          />
          <button
            onClick={onRemove}
            className="absolute top-2 right-2 bg-black/50 hover:bg-black/70 text-white p-1.5 rounded-full transition-colors"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="h-4 w-4"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </button>
        </div>
      )}
    </div>
  );
}

export default ImageUpload;
