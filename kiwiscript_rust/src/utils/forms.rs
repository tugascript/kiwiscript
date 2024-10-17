#[derive(Clone)]
pub struct FormField<T>
where
    T: Clone + ToString,
{
    value: T,
    error: Option<String>,
}

impl<T> FormField<T>
where
    T: Clone + ToString,
{
    pub fn new(value: T) -> Self {
        Self { value, error: None }
    }

    pub fn update(&mut self, value: T) {
        self.value = value;
    }

    pub fn set_error(&mut self, error: String) {
        self.error = Some(error);
    }

    pub fn clear_error(&mut self) {
        self.error = None;
    }

    pub fn val(&self) -> String {
        self.value.to_string()
    }

    pub fn err(&self) -> Option<&String> {
        self.error.as_ref()
    }

    pub fn has_error(&self) -> bool {
        self.error.is_some()
    }
}
