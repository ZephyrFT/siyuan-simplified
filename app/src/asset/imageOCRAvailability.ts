import {isEncryptedBox} from "../util/pathName";

export const getImageOCRAvailability = (path: string, notebookId: string) => {
    const reference = (path || "").split("#", 1)[0];
    const [pathname, query = ""] = reference.split("?", 2);
    const boxId = new URLSearchParams(query).get("box")?.trim();
    const text = !isEncryptedBox(notebookId) && !isEncryptedBox(boxId);
    const local = text && pathname.startsWith("assets/");
    return {text, local};
};
