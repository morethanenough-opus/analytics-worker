
export function assertIsDefined<T>(val: T): asserts val is NonNullable<T> {
  if (val === undefined || val === null) {
    throw new Error(`Expected 'val' to be defined, but received ${val}`);
  }
}


interface UserProfile {
  id: string;
  email: string;
  isActive: boolean;
  roles: string[];
}

