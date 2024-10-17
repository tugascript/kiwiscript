use leptos::*;

pub enum SpinnerColor {
    Primary,
    Secondary,
    Success,
    Danger,
    Warning,
    Info,
    Light,
    Dark,
}

impl ToString for SpinnerColor {
    fn to_string(&self) -> String {
        match self {
            SpinnerColor::Primary => "text-primary".to_string(),
            SpinnerColor::Secondary => "text-secondary".to_string(),
            SpinnerColor::Success => "text-success".to_string(),
            SpinnerColor::Danger => "text-danger".to_string(),
            SpinnerColor::Warning => "text-warning".to_string(),
            SpinnerColor::Info => "text-info".to_string(),
            SpinnerColor::Light => "text-light".to_string(),
            SpinnerColor::Dark => "text-dark".to_string(),
        }
    }
}

#[component]
fn Spinner(role: String, color: SpinnerColor) -> impl IntoView {
    let class = format!("spinner-border {}", color.to_string());
    view! {
        <div class=class role=role>
            <span class="visually-hidden">"Loading..."</span>
        </div>
    }
}
