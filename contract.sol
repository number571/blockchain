pragma solidity ^0.5.0;


contract WorldSkills {

    struct  Estate {
        uint estate_id;
        address owner;
        string info;
        uint squere;
        uint useful_squere;
        address renter_address;
        bool present_status;
        bool sale_status;
        bool rent_status;
    }
    
    struct Present {
        uint estate_id;
        address address_from;
        address address_to;
        bool finished;
    }
    
    struct Sale {
        uint estate_id;
        address owner;
        uint price;
        address payable[] customers;
        uint[] prices;
        bool finished;
    }
    
    struct Rent {
        uint estate_id;
        address payable owner_address;
        address payable renter_address;
        uint time;
        uint money;
        uint deadline;
        bool finished;
    }
    
    Estate[] estates;
    Present[] presents;
    Sale[] sales;
    Rent[] rents;
    
    address admin = msg.sender;
    address payable default_address = 0x0000000000000000000000000000000000000000;
    
    function iam_admin() public view returns(bool) {
        return msg.sender == admin;
    }

    function get_estates_number() public view returns(uint) {
        return estates.length;
    }
    
    function get_presents_number() public view returns(uint) {
        return presents.length;
    }
    
    function get_sales_number() public view returns(uint) {
        return sales.length;
    }
    
    function get_rents_number() public view returns(uint) {
        return rents.length;
    }
    
    function get_estates(uint estate_number) public view returns(uint, address, string memory, uint, uint, address) {
        return(estates[estate_number].estate_id, estates[estate_number].owner, estates[estate_number].info, estates[estate_number].squere, estates[estate_number].useful_squere, estates[estate_number].renter_address);
    }

    function get_estates_statuses(uint estate_number) public view returns(bool, bool, bool) {
        return(estates[estate_number].present_status, estates[estate_number].sale_status, estates[estate_number].rent_status);
    }
    
    function get_presents(uint present_number) public view returns(uint, address, address, bool) {
        return(presents[present_number].estate_id, presents[present_number].address_from, presents[present_number].address_to, presents[present_number].finished);
    }
    
    function get_sales(uint sale_number) public view returns(uint, address, uint,  address payable[] memory, uint[] memory prices, bool) {
        return(sales[sale_number].estate_id, sales[sale_number].owner, sales[sale_number].price, sales[sale_number].customers, sales[sale_number].prices, sales[sale_number].finished);
    }
    
    function get_rents(uint rent_number) public view returns(uint, address, address, uint, uint, uint, bool) {
        return(rents[rent_number].estate_id, rents[rent_number].owner_address, rents[rent_number].renter_address, rents[rent_number].time, rents[rent_number].money, rents[rent_number].deadline, rents[rent_number].finished);
    }
    
    modifier status_OK(uint estate_id) {
        require(estates[estate_id].present_status == false);
        require(estates[estate_id].sale_status == false);
        require(estates[estate_id].rent_status == false);
        _;
    }
    
    modifier is_owner(uint estate_id) {
        require(msg.sender == estates[estate_id].owner);
        _;
    }
    
    modifier is_admin {
        require(msg.sender == admin);
        _;
    }

    function create_estate(address owner, string memory info, uint squere, uint useful_squere) public is_admin{
        estates.push(Estate(estates.length, owner, info, squere, useful_squere, 0x0000000000000000000000000000000000000000, false, false, false));
    }
    
    function create_present(uint estate_id, address address_to) public status_OK(estate_id) is_owner(estate_id) {
        presents.push(Present(estate_id, msg.sender, address_to, false));
        estates[estate_id].present_status = true;
    } 
    
    function cancel_present(uint present_number) payable public {
        require(msg.sender == presents[present_number].address_from);
        require(presents[present_number].finished == false);
        estates[presents[present_number].estate_id].present_status = false;
        presents[present_number].finished = true;
        
    }
    
    function confirm_present(uint present_number) payable public {
        require(msg.sender == presents[present_number].address_to);
        require(presents[present_number].finished == false);
        estates[presents[present_number].estate_id].owner = presents[present_number].address_to;
        estates[presents[present_number].estate_id].present_status = false;
        presents[present_number].finished = true;
        
    }
    
    function create_sale(uint estate_id, uint price) public status_OK(estate_id) is_owner(estate_id){
       address payable[] memory customers;
       uint[] memory prices;
       sales.push(Sale(estate_id, msg.sender, price, customers, prices, false));
       estates[estate_id].sale_status = true;
    }
    
    function cancel_sale(uint sale_number) public {
        require(msg.sender == sales[sale_number].owner);
        require(sales[sale_number].finished == false);
        for (uint i = 0; i < sales[sale_number].customers.length; i++){
            (sales[sale_number].customers[i]).transfer(sales[sale_number].prices[i]);
        }
        estates[sales[sale_number].estate_id].sale_status = false;
        sales[sale_number].finished = true;
        
    }
    
    function check_to_buy(uint sale_number) public payable {
        require(msg.sender != sales[sale_number].owner);
        require(msg.value >= sales[sale_number].price);
        require(sales[sale_number].finished == false);
        uint status = 0;
        for (uint i=0; i < sales[sale_number].customers.length; i++) {
            if (sales[sale_number].customers[i] == msg.sender) {
                status = 1;
                break;
            }
        }
        require(status == 0);
        sales[sale_number].customers.push(msg.sender);
        sales[sale_number].prices.push(msg.value);
    }
    
    function cancel_to_buy(uint sale_number) public payable {
        require(sales[sale_number].finished == false);
        for (uint i=0; i<sales[sale_number].customers.length; i++){
            if (sales[sale_number].customers[i] == msg.sender){
                msg.sender.transfer(sales[sale_number].prices[i]);
                delete sales[sale_number].prices[i];
            }
        }
    }
    
    function confirm_sale(uint sale_number, uint sale_to) public payable {
        require(msg.sender == sales[sale_number].owner);
        require(sales[sale_number].prices[sale_to] != 0);
        require(sales[sale_number].finished == false);
        estates[sales[sale_number].estate_id].owner = sales[sale_number].customers[sale_to];
        msg.sender.transfer(sales[sale_number].prices[sale_to]);
        for (uint i=0; i<sales[sale_number].customers.length; i++){
            if (i != sale_to) {
                sales[sale_number].customers[i].transfer(sales[sale_number].prices[i]);
            }
            else {
                sales[sale_number].prices[sale_to] = 0;
            }
        }
        estates[sales[sale_number].estate_id].sale_status = false;
        sales[sale_number].finished = true;
    }

    function create_rent(uint estate_id, uint time, uint money) public is_owner(estate_id) status_OK(estate_id){
        rents.push(Rent(estate_id, msg.sender, default_address, time, money, 0, false));
        estates[estate_id].rent_status=true;
    }
    
    function to_rent(uint rent_id) public payable{
        require(rents[rent_id].finished == false);
        require(rents[rent_id].renter_address == default_address);
        require(rents[rent_id].owner_address != msg.sender);
        require(rents[rent_id].money == msg.value); 
        rents[rent_id].renter_address = msg.sender;
        estates[rents[rent_id].estate_id].renter_address = msg.sender;
        rents[rent_id].deadline = now + rents[rent_id].time*86400;
        rents[rent_id].owner_address.transfer(rents[rent_id].money);
    }
    
    function cancel_rent(uint rent_id) public {
        require(rents[rent_id].finished == false);
        require(rents[rent_id].owner_address == msg.sender);
        require(rents[rent_id].renter_address == default_address);
        estates[rents[rent_id].estate_id].rent_status=false;
        rents[rent_id].finished = true;
    }
    
    function finish_rent(uint rent_id) public is_owner(rents[rent_id].estate_id) { 
        require(rents[rent_id].finished == false);
        require(rents[rent_id].deadline < now);
        estates[rents[rent_id].estate_id].renter_address = default_address;
        estates[rents[rent_id].estate_id].rent_status=false;
        rents[rent_id].finished = true;
    }
}
