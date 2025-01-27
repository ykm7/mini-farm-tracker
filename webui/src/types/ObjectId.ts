// Remove this line:
// The errors you're encountering are related to the use of Node.js-specific modules (crypto and util) in a browser environment. This is likely caused by the import of ObjectId from 'mongodb', which is typically a server-side library.
// import { ObjectId } from 'mongodb'

// In a new file, e.g., @/types/ObjectId.ts
export type ObjectId = string | { toString(): string };
