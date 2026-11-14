import {clearImageOCRText, copyImageOCRText, openImageOCR} from "../asset/imageOCR";
import {getImageOCRAvailability} from "../asset/imageOCRAvailability";
import {getImageOCRStatus, invalidateImageOCRStatus} from "../asset/imageOCRStatus";
import {fetchPost} from "../util/fetch";
import {IMAGE_OCR_MENU} from "./declarations/imageOCRMenu";
import {createDeclaredMenu} from "./declarations/menuDeclaration";

export const createImageOCRMenu = (image: HTMLElement, notebookID: string): IMenu => {
    const path = image.getAttribute("data-src");
    const availability = getImageOCRAvailability(path, notebookID);
    return createDeclaredMenu(IMAGE_OCR_MENU, {
        textAvailable: availability.text,
        localAvailable: availability.local,
        getStatus: () => getImageOCRStatus(path),
        openResult: () => openImageOCR(image.getAttribute("data-src")),
        copyText: () => { copyImageOCRText(image.getAttribute("data-src")); },
        clearText: () => { void clearImageOCRText(image.getAttribute("data-src")); },
        runLocal: () => {
            const currentPath = image.getAttribute("data-src");
            void fetchPost("/api/asset/ocr", {path: currentPath}).finally(() => invalidateImageOCRStatus(currentPath));
        },
    });
};
