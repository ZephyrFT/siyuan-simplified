import {IMenuDeclaration} from "./menuDeclaration";

export interface IImageOCRMenuContext {
    textAvailable: boolean;
    localAvailable: boolean;
    getStatus: () => Promise<boolean | undefined>;
    openResult: () => void;
    copyText: () => void;
    clearText: () => void;
    runLocal: () => void;
}

export const IMAGE_OCR_MENU: IMenuDeclaration<IImageOCRMenuContext> = {
    id: "ocr",
    label: () => "OCR",
    icon: "iconOCR",
    simple: false,
    type: "entry",
    behavior: context => ({
        ignore: !context.textAvailable,
        bind(element) {
            context.getStatus().then(hasText => {
                if (!element.isConnected || hasText === undefined) {
                    return;
                }
                const localLabel = element.querySelector('[data-id="reOCR"] .b3-menu__label');
                if (localLabel) {
                    localLabel.textContent = hasText ? window.siyuan.languages.reOCR : window.siyuan.languages.performOCR;
                }
            });
        },
    }),
    children: [{
        id: "ocrResult",
        label: () => window.siyuan.languages.editOCRText,
        simple: false,
        type: "entry",
        icon: "iconEdit",
        behavior: context => ({click: context.openResult}),
    }, {
        id: "copyOCRText",
        label: () => window.siyuan.languages.copyOCRText,
        simple: false,
        type: "entry",
        icon: "iconCopy",
        behavior: context => ({click: context.copyText}),
    }, {
        id: "clearOCRText",
        label: () => window.siyuan.languages.clearOCRText,
        simple: false,
        type: "entry",
        icon: "iconTrashcan",
        behavior: context => ({click: context.clearText}),
    }, {
        id: "separator_reOCR",
        label: () => "",
        simple: true,
        type: "separator",
        behavior: context => ({ignore: !context.localAvailable}),
    }, {
        id: "reOCR",
        label: () => window.siyuan.languages.performOCR,
        simple: false,
        type: "entry",
        icon: "iconOCR",
        behavior: context => ({ignore: !context.localAvailable, click: context.runLocal}),
    }],
};
