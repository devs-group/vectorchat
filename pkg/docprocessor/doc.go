// Package docprocessor provides document processing capabilities including
// file conversion to markdown, text chunking, and metadata extraction.
//
// The package is designed to handle various file formats and convert them
// into searchable chunks suitable for vector databases and RAG systems.
//
// Main Components:
//   - Processor: Main processing engine for files and text
//   - Chunker: Advanced text and markdown chunking with semantic awareness
//   - Utils: Helper functions for file handling and metadata generation
//
// Example Usage:
//
//	// Create a processor with markitdown client
//	markitdownClient, err := NewMarkitdownClient(apiURL)
//	if err != nil {
//		log.Fatal(err)
//	}
//	processor := NewProcessor(markitdownClient)
//
//	// Process a file upload
//	result, err := processor.ProcessFile(ctx, fileHeader)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Process plain text
//	textResult, err := processor.ProcessText("Your text content here")
//	if err != nil {
//		log.Fatal(err)
//	}
//
// Supported Operations:
//   - File format conversion via MarkItDown service
//   - Intelligent markdown chunking with section awareness
//   - Text chunking with configurable sizes
//   - Metadata wrapping for vector storage
//   - File validation with hardcoded supported extensions
package docprocessor
