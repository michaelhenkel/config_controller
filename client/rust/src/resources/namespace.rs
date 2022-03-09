use crate::resources::traits::{ProcessResource};
use config_client::protos::k8s::io::api::core::v1;

impl ProcessResource for v1::Namespace {
    fn kind(&self) -> String { "RoutingInstance".to_string() }
    fn add(&self) { 
        println!("add for {} {}/{}",self.kind(), self.metadata.as_ref().unwrap().namespace(), self.metadata.as_ref().unwrap().name())
    }
    fn update(&self) { 
        println!("update for {} {}/{}",self.kind(), self.metadata.as_ref().unwrap().namespace(), self.metadata.as_ref().unwrap().name())
    }
    fn delete(&self) { 
        println!("delete for {} {}/{}",self.kind(), self.metadata.as_ref().unwrap().namespace(), self.metadata.as_ref().unwrap().name())
    }
}