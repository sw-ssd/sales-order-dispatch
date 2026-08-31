import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

/** 合併 Tailwind class 字串,衝突類別以後者為準(全站共用 helper,暫置 features/users)。 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}
