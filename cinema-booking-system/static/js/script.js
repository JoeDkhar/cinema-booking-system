// Debug logging to verify script is loading
console.log("Cinema booking system script loaded");

document.addEventListener('DOMContentLoaded', function() {
    console.log("DOM content loaded event triggered");
    
    // Handle image load failures
    const posterImages = document.querySelectorAll('.movie-poster img');
    console.log("Found poster images:", posterImages.length);
    
    posterImages.forEach(img => {
        console.log("Setting error handler for image:", img.src);
        
        img.onerror = function() {
            console.log("Image load error for:", this.src);
            
            // Try local path if external URL fails
            if (this.src.startsWith('https://')) {
                const filename = this.src.split('/').pop();
                console.log("Trying local path:", '/static/images/' + filename);
                this.src = '/static/images/' + filename;
            } else {
                // If local path also fails, use placeholder
                console.log("Using placeholder image");
                this.src = '/static/images/movie-placeholder.jpg';
                this.alt = 'Movie poster not available';
            }
        };
    });
    
    // Add feature detection to verify JavaScript is running
    const featureDetectionElement = document.createElement('div');
    featureDetectionElement.id = 'js-feature-detection';
    featureDetectionElement.style.display = 'none';
    document.body.appendChild(featureDetectionElement);
});