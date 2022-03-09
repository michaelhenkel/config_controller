use config_client::protos::github::com::michaelhenkel::config_controller::pkg::apis::v1;

pub mod traits;
pub mod virtualmachineinterface;
pub mod virtualnetwork;
pub mod virtualrouter;
pub mod virtualmachine;
pub mod routinginstance;
pub mod instanceip;
pub mod namespace;

pub fn get_resource(response: v1::Response) -> Box<dyn traits::ProcessResource>{
    match response.new.unwrap().resource.unwrap() {
        v1::resource::Resource::VirtualNetwork(res) => Box::new(res),
        v1::resource::Resource::VirtualMachineInterface(res) =>  Box::new(res),
        v1::resource::Resource::VirtualRouter(res) =>  Box::new(res),
        v1::resource::Resource::InstanceIp(res) => Box::new(res),
        v1::resource::Resource::VirtualMachine(res) => Box::new(res),
        v1::resource::Resource::RoutingInstance(res) => Box::new(res),
        v1::resource::Resource::Namespace(res) => Box::new(res),
    }
}