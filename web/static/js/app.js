(() => {
    const root = document.documentElement;
    const themeToggle = document.getElementById("theme-toggle");
    const themeLabel = themeToggle?.querySelector(".theme-label");
    const savedTheme = localStorage.getItem("pinharbor-theme");
    const initialTheme = savedTheme === "light" ? "light" : "dark";

    const applyTheme = (theme) => {
        root.dataset.theme = theme;
        const isDark = theme === "dark";
        if (themeToggle) {
            themeToggle.setAttribute("aria-pressed", String(isDark));
            themeToggle.setAttribute("aria-label", isDark ? "Switch to light mode" : "Switch to dark mode");
            themeToggle.firstElementChild.textContent = isDark ? "☾" : "☀";
        }
        if (themeLabel) {
            themeLabel.textContent = isDark ? "Dark" : "Light";
        }
    };

    applyTheme(initialTheme);
    themeToggle?.addEventListener("click", () => {
        const nextTheme = root.dataset.theme === "dark" ? "light" : "dark";
        localStorage.setItem("pinharbor-theme", nextTheme);
        applyTheme(nextTheme);
    });

    const presetSelect = document.getElementById("run-mode");
    const presetSummary = document.getElementById("preset-summary");
    const depth = document.getElementById("depth");
    const maxTime = document.getElementById("maxtime");
    const fastMode = document.getElementById("fastmode");
    const email = document.getElementById("email");
    const extraReviews = document.getElementById("extra-reviews");
    const advancedNote = document.querySelector("#advanced-settings .summary-note");
    const controlled = Array.from(document.querySelectorAll(".preset-controlled input"));

    const presets = {
        auto: {
            title: "Auto complete",
            description: "No clock cutoff · full result feed · email discovery · stops when all discovered work is finished.",
            depth: "0",
            maxTime: "",
            fastMode: false,
            email: true,
            extraReviews: false,
        },
        quick: {
            title: "Quick scan",
            description: "15-minute cap · 25 scrolls · core business details · email discovery off.",
            depth: "25",
            maxTime: "15m",
            fastMode: false,
            email: false,
            extraReviews: false,
        },
        balanced: {
            title: "Balanced",
            description: "1-hour cap · 100 scrolls · email discovery on · suitable for everyday lists.",
            depth: "100",
            maxTime: "1h",
            fastMode: false,
            email: true,
            extraReviews: false,
        },
        deep: {
            title: "Deep collection",
            description: "3-hour cap · 250 scrolls · email discovery and extra reviews enabled.",
            depth: "250",
            maxTime: "3h",
            fastMode: false,
            email: true,
            extraReviews: true,
        },
        custom: {
            title: "Custom controls",
            description: "Set depth, enrichment, Fast Mode, and a maximum time manually.",
        },
    };

    const applyPreset = () => {
        if (!presetSelect) return;
        const name = presetSelect.value;
        const preset = presets[name] || presets.auto;
        const custom = name === "custom";

        if (!custom) {
            depth.value = preset.depth;
            maxTime.value = preset.maxTime;
            fastMode.checked = preset.fastMode;
            email.checked = preset.email;
            extraReviews.checked = preset.extraReviews;
        }

        controlled.forEach((field) => {
            field.disabled = !custom;
        });
        document.getElementById("advanced-settings")?.classList.toggle("preset-locked", !custom);
        if (advancedNote) advancedNote.textContent = custom ? "Manual" : "Preset managed";
        if (presetSummary) {
            presetSummary.innerHTML = "";
            const title = document.createElement("strong");
            title.textContent = preset.title;
            const description = document.createElement("span");
            description.textContent = preset.description;
            presetSummary.append(title, description);
        }
    };

    presetSelect?.addEventListener("change", applyPreset);
    applyPreset();

    document.getElementById("reset-form")?.addEventListener("click", () => {
        const form = document.querySelector("form[hx-post='/scrape']");
        form?.reset();
        if (presetSelect) presetSelect.value = "auto";
        applyPreset();
        const error = document.getElementById("error-container");
        if (error) error.textContent = "";
        document.getElementById("name")?.focus();
    });
})();
