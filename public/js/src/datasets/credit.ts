export class Credit {
    BottomText: string
    MajorCategory: string
    MinorCategory: string
    Name: string
    Order: string
    Project: string
    static fromJSON = (obj: Object): Credit => obj as Credit
}
