import { useCallback, useEffect, useState } from "react";

import Button from "react-bootstrap/Button";
import ButtonGroup from "react-bootstrap/ButtonGroup";
import FormControl from "react-bootstrap/FormControl";
import { ShowHideToggle } from "./ShowHideToggle";
import { isEmpty } from "lodash/fp";

export function useMenuSelection<T extends { Name: string }>(
    choices: T[],
    pathMatcher: RegExp,
    permaLink: (x: T) => string,
    defaultChoice?: T
): {
    selected: T | undefined;
    onSelect: (selection: T) => void;
} {
    const [selected, setSelected] = useState<T | undefined>(defaultChoice);

    window.onpopstate = (event) => {
        if (event.state) setSelected(event.state);
    };

    const onSelect = useCallback((selection: T, shouldAddState = true) => {
        setSelected(selection);
        if (shouldAddState) {
            window.history.pushState(selection, "", permaLink(selection));
        } else {
            window.history.replaceState(selection, "", permaLink(selection));
        }
    }, []);

    useEffect(() => {
        let want = defaultChoice;

        const params = new URLSearchParams(document.location.search);
        if (!isEmpty(params.get("name"))) {
            want = choices.filter((p) => p.Name == params.get("name")).pop();
        }

        const pathMatch = document.location.pathname.match(pathMatcher);
        if (pathMatch && pathMatch.length == 2 && pathMatch[1].length > 0) {
            want = choices.filter((p) => p.Name == pathMatch[1]).pop();
        }
        if (want) {
            onSelect(want, false);
        } else if (defaultChoice) {
            onSelect(defaultChoice, false);
        }
    }, [defaultChoice]);

    return {
        selected,
        onSelect,
    };
}

type ToggleParams<T> = {
    hidden: boolean;
    title: string;
    filter: (on: boolean, x: T) => boolean;
};

export function FancyProjectMenu<T extends { Name: string }>(props: {
    choices: T[];
    permaLink: (x: T) => string;
    selected: T | undefined;
    onSelect: (selection: T) => void;
    buttonContent: (x: T) => JSX.Element;
    searchChoices?: (q: string, choices: T[]) => T[];
    toggles?: Array<ToggleParams<T>>;
}) {
    const [searchInput, setSearchInput] = useState("");
    const menuToggles = useToggles(props.toggles ?? []);

    const searcher = props.searchChoices ?? (() => props.choices);
    const filter = (x: T) =>
        isEmpty(menuToggles) || menuToggles.map((t) => t.filter(t.state, x)).reduce((a, b) => a && b);
    const wantChoices = searcher(searchInput, props.choices).filter(filter);

    const onClickChoice = (want: T) => {
        props.onSelect(want);
    };

    return (
        <div>
            <MenuToggles toggles={menuToggles} />
            <SearchBox hidden={props.searchChoices == null} setSearchInput={setSearchInput} />
            <div className="d-grid">
                <ButtonGroup vertical className="m-2">
                    {wantChoices.map((want) => (
                        <Button
                            variant={props.selected && props.selected.Name == want.Name ? "light" : "outline-light"}
                            key={want.Name}
                            onClick={() => onClickChoice(want)}
                        >
                            {props.buttonContent(want)}
                        </Button>
                    ))}
                </ButtonGroup>
            </div>
        </div>
    );
}

export type Toggle<T> = {
    hidden: boolean;
    title: string;
    state: boolean;
    setState: (b: boolean) => void;
    filter: (on: boolean, x: T) => boolean;
};

export function useToggles<T>(toggles: ToggleParams<T>[]): Toggle<T>[] {
    const [toggleState, setToggleState] = useState(0);
    return toggles.map((t, i) => ({
        ...t,
        // prettier-ignore
        state: (toggleState & (1 << i)) == (1 << i),
        setState: (x: boolean) => setToggleState(x ? toggleState | (1 << i) : toggleState & ~(1 << i)),
    }));
}

export function MenuToggles<T>(props: { toggles: Toggle<T>[] }): JSX.Element {
    const toggles = props.toggles.filter((t) => !t.hidden);
    return isEmpty(props.toggles) ? (
        <div />
    ) : (
        <div className={"d-flex flex-row justify-content-center"}>
            {toggles.map((toggle) => (
                <ShowHideToggle
                    key={toggle.title}
                    title={toggle.title}
                    state={toggle.state}
                    setState={toggle.setState}
                />
            ))}
        </div>
    );
}

export function SearchBox(props: { hidden: boolean; setSearchInput: (q: string) => void }): JSX.Element {
    return props.hidden ? (
        <div />
    ) : (
        <div className="d-flex flex-row justify-content-center">
            <FormControl
                className="m-2"
                placeholder="search projects"
                onChange={(event) => props.setSearchInput(event.target.value.toLowerCase())}
            />
        </div>
    );
}
