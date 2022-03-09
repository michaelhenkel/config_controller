pub trait ProcessResource {
    fn kind(&self) -> String;
    fn add(&self);
    fn update(&self);
    fn delete(&self);
}